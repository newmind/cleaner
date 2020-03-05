package main // import "gitlab.markany.com/argos/cleaner"

import (
	"container/list"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.markany.com/argos/cleaner/fileinfo"
	"gitlab.markany.com/argos/cleaner/scanner"
	"gitlab.markany.com/argos/cleaner/watcher"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var appName = "cleaner"

var (
	// configuration
	deleteEmptyDir bool
	deleteHidden   bool // . 파일/디렉토리 삭제할지, (default : false)
	interval       string
	freePercent    int // 여유공간 몇퍼센트 유지할지
	dryRun         bool
	debug          bool
)

func init() {
	// Read command line flags
	flag.BoolVar(&deleteEmptyDir, "delete-empty-dir", true, "Delete if dir is empty")
	flag.BoolVar(&deleteHidden, "delete-hidden", false, "Delete .(dot) files or dirs")
	flag.StringVar(&interval, "interval", "100ms", "Deletor poll interval")
	flag.IntVar(&freePercent, "free-percent", 10, "Keep free percent")
	flag.BoolVar(&dryRun, "dry-run", true, "Dry run, doesn't remove files if true")
	flag.BoolVar(&debug, "debug", true, "Debug mode")

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	})
	if debug {
		log.SetLevel(log.DebugLevel)
	}

}

func loadConfig() {
	dirs := flag.Args()

	viper.Set("delete-empty-dir", deleteEmptyDir)
	viper.Set("delete-hidden", deleteHidden)

	// Parse the interval string into a time.Duration.
	parsedInterval, err := time.ParseDuration(interval)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}
	viper.Set("interval", parsedInterval)
	viper.Set("free-percent", freePercent)
	viper.Set("debug", debug)

	if len(dirs) == 0 {
		fmt.Printf("Usage : ./%s [options] path ...  \n", appName)
		flag.PrintDefaults()
		os.Exit(1)
		//curDir, err := os.Getwd()
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//// 아무런 인자가 없다면, temp디렉토리를 추가
		//dirs = append(dirs, os.TempDir())
	}
	// directory 에 현재 실행파일의  디렉토리가 실수로 포함되지 않게 함
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(executable)
	for _, d := range dirs {
		if d == exPath {
			fmt.Printf("Path must not directory of executable")
			fmt.Printf("Usage : ./%s [options] path ...  \n", appName)
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	viper.Set("dirs", dirs)
}

func main() {
	flag.Parse()
	loadConfig()

	log.Infof("Starting %v ...\n", appName)

	//
	// 파일 감지를 먼저 시작하고, 스캔을 나중에 함. 파일이 중복되어도 큰 문제 없음
	//

	// 1. 파일감지 초기화
	// 파일정보 전송용 버퍼 채널
	// 감시된 파일은 생성(created)된 순서대로 들어오므로, linked queue 사용
	log.Info("Notification handler starting ...")
	qFilesWatched := list.New()
	mutexQ := sync.Mutex{}
	chWatcher := make(chan fileinfo.FileInfo, 200)

	// 2. 파일와처 채널 리시버
	go watcher.HandleFileCreation(chWatcher, qFilesWatched, &mutexQ)

	// 3. 디렉토리의 새로운 파일 감시
	for _, dir := range viper.GetStringSlice("dirs") {
		log.Infof("Watching directory \"%v\" ...", dir)
		if err := watcher.Watch(dir, chWatcher); err != nil {
			log.Panic(err)
		}
	}

	// 4. 디렉토리 스캔
	log.Infof("File creation handler starting [%s] ...", viper.GetStringSlice("dirs"))
	scannedFiles := scanner.ScanAllFiles(viper.GetStringSlice("dirs"))

	// 5. 여유 공간 확보를 위해서, 오래된 파일부터 삭제
	usage, err := disk.Usage(viper.GetStringSlice("dirs")[0])
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Disk usage : ", usage)
	log.Info("Free up disk ...")
	go freeUpSpace(scannedFiles, qFilesWatched, mutexQ)

	done := make(chan bool)
	<-done
}

func freeUpSpace(scannedFiles []*fileinfo.FileInfo, qFilesWatched *list.List, mutexQ sync.Mutex) {
	pollingInterval := viper.GetDuration("interval")
	dirs := viper.GetStringSlice("dirs")
	freePercent := viper.GetInt("free-percent")

	for {
		//TODO: disk 별로 여유공간 유지해야함
		usage, err := disk.Usage(dirs[0])
		if err != nil {
			log.Fatal(err)
			panic(err)
		}

		// usage.Total 에는 사스템 예약공간이 포함되어서 용량이 다를수 있음
		// https://github.com/shirou/gopsutil/issues/562
		total := usage.Used + usage.Free
		deletedSize := int64(0)
		usedPercent := usage.UsedPercent
		for usedPercent+float64(freePercent) >= 100 {
			// 1. scan 한 파일 먼저 지우고,(다 지웠으면)
			// 2. 파일감지로 새로 추가된 파일들 지움
			if len(scannedFiles) > 0 {
				last := scannedFiles[len(scannedFiles)-1]
				if fi, err := os.Lstat(last.Path); err == nil {
					deletedSize += fi.Size()
					if err := remove(last.Path); err != nil {
						log.Error(err)
					}
					log.Debug("Deleted :", last.Path)
					if deleteEmptyDir {
						// try removing dir if empty
						err := remove(filepath.Dir(last.Path))
						if err == nil {
							log.Debug("Deleted[dir] :", filepath.Dir(last.Path))
						}
					}
				}
				// remove from slice
				scannedFiles = scannedFiles[:len(scannedFiles)-1]
			} else if qFilesWatched.Len() > 0 {
				mutexQ.Lock()
				if elem := qFilesWatched.Front(); elem != nil {
					path := elem.Value.(*fileinfo.FileInfo).Path
					if fi, err := os.Lstat(path); err == nil {
						deletedSize += fi.Size()
						remove(path)
						log.Debug("Deleted :", path)
						if deleteEmptyDir {
							// try removing dir if empty
							err := remove(filepath.Dir(path))
							if err == nil {
								log.Debug("Deleted[dir] :", filepath.Dir(path))
							}
						}
					}
					qFilesWatched.Remove(elem)
				}
				mutexQ.Unlock()
			}

			// 다시 계산
			usedPercent = float64((usage.Used - uint64(deletedSize)) / total)
		}

		time.Sleep(pollingInterval)
	}
}

func remove(path string) error {
	dryRun := viper.GetBool("dry-run")
	if dryRun {
		return nil
	}
	return os.Remove(path)
}
