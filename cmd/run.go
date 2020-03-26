package cmd

import (
	"container/list"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"gitlab.markany.com/argos/cleaner/service"

	"gitlab.markany.com/argos/cleaner/fileinfo"
	"gitlab.markany.com/argos/cleaner/scanner"
	"gitlab.markany.com/argos/cleaner/watcher"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appName = "cleaner"

var (
	// configuration
	deleteEmptyDir bool
	deleteHidden   bool // dot(.) 파일/디렉토리 삭제할지, (default : false)
	interval       string
	freePercent    int // 여유공간 몇퍼센트 유지할지
	dryRun         bool
	debug          bool
	paths          []string
	serverPort     string

	cpuprofile string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run cleaner",
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
		run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().StringSliceVar(&paths, "paths", []string{}, "paths to watch and clean (required)")
	runCmd.Flags().BoolVar(&deleteEmptyDir, "delete_empty_dir", true, "delete empty dir")
	runCmd.Flags().BoolVar(&deleteHidden, "delete_hidden", false, "delete .(dot) files or dirs")
	runCmd.Flags().StringVar(&interval, "interval", "100ms", "poll interval to check free-space")
	runCmd.Flags().IntVar(&freePercent, "free_percent", 10, "Keep free percent")
	runCmd.Flags().BoolVar(&dryRun, "dry_run", true, "dry run")
	runCmd.Flags().BoolVar(&debug, "debug", true, "use debug logging mode")
	runCmd.Flags().StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	runCmd.Flags().StringVar(&serverPort, "server_port", "8089", "http port")
}

func loadConfig() {
	// Parse the interval string into a time.Duration.
	parsedInterval, err := time.ParseDuration(interval)
	if err != nil {
		log.Error(err)
		fmt.Println(`Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)
		os.Exit(1)
	}
	viper.Set("interval", parsedInterval)

	dirs := viper.GetStringSlice("paths")

	if len(dirs) == 0 {
		fmt.Printf("Usage : ./%s run [options] --paths=/foo,/bar  \n", appName)
		//flag.PrintDefaults()
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
			fmt.Printf("Usage : ./%s run [options] --paths=/foo,/bar  \n", appName)
			//flag.PrintDefaults()
			os.Exit(1)
		}
	}

	for i, d := range dirs {
		d := filepath.Clean(d)
		if realD, err := filepath.EvalSymlinks(d); err != nil || realD != d {
			fmt.Printf("Symbolic link is not supported : %s -> %s \n", d, realD)
			os.Exit(1)
		}
		dirs[i] = filepath.Clean(d)
	}
	viper.Set("paths", dirs)
	fmt.Println("config :", viper.AllSettings())
}

func initLogger() {
	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	if viper.GetBool("debug") {
		log.SetReportCaller(true)
		formatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
			// shorten filename and remove func-name
			return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		}
		log.SetLevel(log.DebugLevel)
	}
}

func formatFilePath(path string) string {
	arr := strings.Split(path, "/")
	return arr[len(arr)-1]
}

func run() {
	loadConfig()
	initLogger()

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.Infof("Starting %v (debug=%v, dryRun=%v)...\n", appName, debug, dryRun)

	//
	// 파일 감지를 먼저 시작하고, 스캔을 나중에 함. 파일이 중복되어도 큰 문제 없음
	//

	// 1. 파일감지 초기화
	// 파일정보 전송용 버퍼 채널
	// 감시된 파일은 생성(created)된 순서대로 들어오므로, linked queue 사용
	log.Info("Notification handler starting ...")
	qFilesWatched := list.New()
	mutexQ := sync.Mutex{}
	chWatcher := make(chan fileinfo.FileInfo, 1000)

	// 2. 파일와처 채널 리시버
	go watcher.HandleFileCreation(chWatcher, qFilesWatched, &mutexQ)

	// 3. 디렉토리의 새로운 파일 감시
	for _, dir := range viper.GetStringSlice("paths") {
		log.Infof("Watching a directory \"%v\" ...", dir)
		if err := watcher.Watch(dir, chWatcher); err != nil {
			log.Panic(err)
		}
	}

	// 4. 디렉토리 스캔
	var scannedFiles []*fileinfo.FileInfo
	log.Infof("Scanning directories [%s] ...", viper.GetStringSlice("paths"))
	scannedFiles = scanner.ScanAllFiles(viper.GetStringSlice("paths"))
	log.Infof("  scanned %v files\n", len(scannedFiles))

	// 5. 여유 공간 확보를 위해서, 오래된 파일부터 삭제
	usage, err := disk.Usage(viper.GetStringSlice("paths")[0])
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Disk usage : ", usage)
	log.Info("Free up disk ...")
	go freeUpSpace(scannedFiles, qFilesWatched, mutexQ)

	go service.StartWebServer(viper.GetString("server_port"))

	done := make(chan bool)
	handleSigterm(func() {
		log.Info("Exit")
		if cpuprofile != "" {
			pprof.StopCPUProfile()
		}
		service.StopWebServer()
		done <- true
	})
	<-done
}

func freeUpSpace(scannedFiles []*fileinfo.FileInfo, qFilesWatched *list.List, mutexQ sync.Mutex) {
	pollingInterval := viper.GetDuration("interval")
	dirs := viper.GetStringSlice("paths")
	freePercent := viper.GetInt("free_percent")

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
			usedPercent = (float64(usage.Used-uint64(deletedSize)) / float64(total)) * 100.0
		}

		time.Sleep(pollingInterval)
	}
}

func remove(path string) error {
	dryRun := viper.GetBool("dry_run")
	if dryRun {
		return nil
	}
	return os.Remove(path)
}

// Handles Ctrl+C or most other means of "controlled" shutdown gracefully. Invokes the supplied func before exiting.
func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		// os.Exit(1)
	}()
}
