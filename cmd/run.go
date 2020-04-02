package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"gitlab.markany.com/argos/cleaner/common"
	"gitlab.markany.com/argos/cleaner/diskinfo"
	"gitlab.markany.com/argos/cleaner/service"
	"gitlab.markany.com/argos/cleaner/vods"

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
	serverPort     string
	vodPath        string
	imagePath      string

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
	runCmd.Flags().BoolVar(&deleteEmptyDir, "delete_empty_dir", true, "delete empty dir")
	runCmd.Flags().BoolVar(&deleteHidden, "delete_hidden", false, "delete .(dot) files or dirs")
	runCmd.Flags().StringVar(&interval, "interval", "1m", "poll interval to check free-space")
	runCmd.Flags().IntVar(&freePercent, "free_percent", 10, "Keep free percent")
	runCmd.Flags().BoolVar(&dryRun, "dry_run", true, "dry run")
	runCmd.Flags().BoolVar(&debug, "debug", true, "use debug logging mode")
	runCmd.Flags().StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	runCmd.Flags().StringVar(&serverPort, "server_port", "8089", "http port")
	runCmd.Flags().StringVar(&vodPath, "vod_path", "", "vod path (required)")
	runCmd.Flags().StringVar(&imagePath, "image_path", "", "image path (required)")
}

func loadConfig() {
	// Parse the interval string into a time.Duration.
	_, err := time.ParseDuration(viper.GetString("INTERVAL"))
	if err != nil {
		log.Error(err)
		fmt.Println(`Valid time units are "s", "m", "h"`)
		os.Exit(1)
	}

	dirs := []string{}
	if len(viper.GetString("VOD_PATH")) > 0 {
		vodPath = filepath.Clean(viper.GetString("VOD_PATH"))
		dirs = append(dirs, vodPath)
	}
	if len(viper.GetString("IMAGE_PATH")) > 0 {
		imagePath = filepath.Clean(viper.GetString("IMAGE_PATH"))
		dirs = append(dirs, imagePath)
	}

	if len(dirs) == 0 {
		fmt.Printf("Usage : ./%s run [options] --vod_path=/foo --image_path=/images \n", appName)
		fmt.Println("  둘중에 하나는 있어야 함")
		os.Exit(1)
	}

	// directory 에 현재 실행파일의  디렉토리가 실수로 포함되지 않게 함
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(executable)
	for _, d := range dirs {
		d := filepath.Clean(d)
		d, err = filepath.Abs(d)
		if strings.HasPrefix(exPath, d) {
			fmt.Printf("Path must not directory of executable\n")
			fmt.Printf("Usage : ./%s run [options] --vod_path=/foo --image_path=/images\n", appName)
			os.Exit(1)
		}
	}

	for _, d := range dirs {
		d := filepath.Clean(d)
		if _, err := os.Stat(d); err == nil {
			if realD, err := filepath.EvalSymlinks(d); err != nil || realD != d {
				fmt.Printf("Symbolic link is not supported : %s -> %s \n", d, realD)
				os.Exit(1)
			}
		}
	}
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

type PathType int

const (
	PathTypeVOD PathType = iota
	PathTypeImage
)

type PathInfo struct {
	Path string
	Type PathType
}

func (p PathInfo) String() string {
	if p.Type == PathTypeVOD {
		return p.Path + "[VOD]"
	} else {
		return p.Path + "[image]"
	}
}

func run() {
	loadConfig()
	initLogger()

	if viper.GetString("CPUPROFILE") != "" {
		f, err := os.Create(viper.GetString("CPUPROFILE"))
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.Infof("Starting %v (debug=%v, dryRun=%v)...", appName, viper.GetBool("DEBUG"), viper.GetBool("DRY_RUN"))

	log.Info("All partitions : ")
	partitions, err := diskinfo.GetAllPartitions()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range partitions {
		log.Info(p.String())
	}

	// map[partition] : []string{path}
	diskMap := getDiskPathMap(
		PathInfo{Path: vodPath, Type: PathTypeVOD},
		PathInfo{Path: imagePath, Type: PathTypeImage})

	cronCleaner := cron.New(cron.WithSeconds())

	for partition, pathInfos := range diskMap {
		log.Infof("Scheduled to delete %s %s", partition, pathInfos)
		isRunning := &common.TAtomBool{}
		p := partition  // capture
		pi := pathInfos // capture
		deleter := func() {
			freeUpDisk(p, pi, isRunning)
		}

		_, err := cronCleaner.AddFunc(fmt.Sprintf("@every %s", viper.GetString("INTERVAL")), deleter)
		if err != nil {
			log.Fatal(err)
		}
	}
	cronCleaner.Start()

	go service.StartWebServer(viper.GetString("SERVER_PORT"))

	done := make(chan bool)
	handleSigterm(func() {
		log.Info("Exit")
		cronCleaner.Stop()
		if viper.GetString("CPUPROFILE") != "" {
			pprof.StopCPUProfile()
		}
		service.StopWebServer()
		done <- true
	})
	<-done
}

func getDiskPathMap(paths ...PathInfo) map[string][]PathInfo {
	diskMap := map[string][]PathInfo{}

	for _, pathInfo := range paths {
		if len(pathInfo.Path) > 0 {
			mountPoint := diskinfo.GetMountpoint(pathInfo.Path)
			if len(mountPoint) == 0 {
				log.Fatalln("Could not find mountpoint of ", pathInfo)
			}
			log.Infof("Mountpoint of '%s' is '%s'", pathInfo, mountPoint)
			if val, ok := diskMap[mountPoint]; ok {
				diskMap[mountPoint] = append(val, pathInfo)
			} else {
				diskMap[mountPoint] = []PathInfo{pathInfo}
			}
		}
	}
	return diskMap
}

func freeUpDisk(partition string, pathInfos []PathInfo, isRunning *common.TAtomBool) {
	log.Debugln("Free up disk", partition, pathInfos)
	if isRunning.Get() {
		log.Warnln("still running ...")
		return
	}
	isRunning.Set(true)
	defer isRunning.Set(false)

	freePercent := viper.GetInt("FREE_PERCENT")
	dryRun := viper.GetBool("DRY_RUN")

	usage, err := disk.Usage(partition)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	var allVodList []*vods.VodInfo = nil
	var allImageList []*vods.ImageInfo = nil

	for _, info := range pathInfos {
		switch info.Type {
		case PathTypeVOD:
			allVodList = vods.ListAllVODs(info.Path)
		case PathTypeImage:
			allImageList = vods.ListAllImages(info.Path)
		}
	}

	for usage.UsedPercent+float64(freePercent) >= 100 {

		var oldVodInfos []*vods.VodInfo = nil
		var oldImageInfos []*vods.ImageInfo = nil

		for _, info := range pathInfos {
			switch info.Type {
			case PathTypeVOD:
				oldVodInfos = vods.ListOldestCCTV(allVodList)
			case PathTypeImage:
				oldImageInfos = allImageList
			}
		}

		var foundVod bool
		var yVod, mVod, dVod int
		if len(oldVodInfos) > 0 {
			foundVod, yVod, mVod, dVod = oldVodInfos[0].GetOldestDay()
		}

		var foundImage bool
		var yImage, mImage, dImage int
		if len(oldImageInfos) > 0 {
			foundImage, yImage, mImage, dImage = oldImageInfos[0].GetOldestDay()
		}

		if foundVod && foundImage {
			if yVod < yImage ||
				yVod <= yImage && mVod < mImage ||
				yVod <= yImage && mVod <= mImage || dVod < dImage {
				oldVodInfos[0].DeleteOldestDay(!dryRun)
			} else {
				oldImageInfos[0].DeleteOldestDay(!dryRun)
			}
		} else if foundVod {
			oldVodInfos[0].DeleteOldestDay(!dryRun)
		} else if foundImage {
			oldImageInfos[0].DeleteOldestDay(!dryRun)
		} else {
			log.Warnf("Could not free up disk [%s]", partition)
			break
		}

		// 다시 계산
		usage, err = disk.Usage(partition)
		if err != nil {
			log.Fatal(err)
		}
	}
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
