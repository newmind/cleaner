package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime/pprof"

	"gitlab.markany.com/argos/cleaner/fileinfo"
	"gitlab.markany.com/argos/cleaner/scanner"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const appName = "deletor"

var (
	path           string
	deleteEmptyDir bool
	lstat          bool
	dryRun         bool
	cpuprofile     string
)

func init() {
	flag.StringVar(&path, "path", "", "Path to delete")
	flag.BoolVar(&deleteEmptyDir, "delete_empty_dir", true, "Delete if dir is empty")
	flag.BoolVar(&lstat, "lstat", true, "call lstat() every remove")
	flag.BoolVar(&dryRun, "dry_run", true, "Dry run, doesn't remove files if true")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "CPU profile")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	viper.Set("dry_run", dryRun)

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Infof("Starting %s...", appName)

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var scannedFiles []*fileinfo.FileInfo
	log.Infof("Scanning directories [%s] ...", path)
	scannedFiles = scanner.ScanAllFiles([]string{path})
	log.Infof("  scanned %v files", len(scannedFiles))

	usage, err := disk.Usage(path)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	total := usage.Used + usage.Free
	deletedSize := int64(0)
	usedPercent := usage.UsedPercent
	freePercent := 10
	for usedPercent+float64(freePercent) >= 100 {
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
		} else {
			break
		}

		// 다시 계산
		usedPercent = (float64(usage.Used-uint64(deletedSize)) / float64(total)) * 100.0
		//log.Debugf("%v = (%v - %v) / %v", usedPercent, usage.Used, deletedSize, total)
	}
	log.Infof("Total size deleted %d", deletedSize)

	log.Info("Exit")
}

func remove(path string) error {
	dryRun := viper.GetBool("dry_run")
	if dryRun {
		return nil
	}
	return os.Remove(path)
}
