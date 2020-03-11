package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime/pprof"

	"gitlab.markany.com/argos/cleaner/fileinfo"
	"gitlab.markany.com/argos/cleaner/scanner"

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
	flag.BoolVar(&deleteEmptyDir, "delete-empty-dir", true, "Delete if dir is empty")
	flag.BoolVar(&lstat, "lstat", true, "call lstat() every remove")
	flag.BoolVar(&dryRun, "dry-run", true, "Dry run, doesn't remove files if true")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "CPU profile")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

	viper.Set("dryRun", dryRun)

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
	log.Infof("  scanned %v files\n", len(scannedFiles))

	var deletedSize int64

	for len(scannedFiles) > 0 {
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
	}
	log.Infof("Total size deleted %d", deletedSize)

	log.Info("Exit")
}

func remove(path string) error {
	dryRun := viper.GetBool("dry-run")
	if dryRun {
		return nil
	}
	return os.Remove(path)
}
