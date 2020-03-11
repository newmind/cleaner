package main

import (
	"flag"
	"os"
	"runtime/pprof"

	"gitlab.markany.com/argos/cleaner/fileinfo"
	"gitlab.markany.com/argos/cleaner/scanner"

	log "github.com/sirupsen/logrus"
)

const appName = "deletor"

var (
	path       string
	lstat      bool
	cpuprofile string
)

func init() {
	flag.StringVar(&path, "path", "", "Path to delete")
	flag.BoolVar(&lstat, "lstat", false, "call lstat() every remove")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "CPU profile")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(1)
	}

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

	for _, fi := range scannedFiles {
		if lstat {
			_, _ = os.Lstat(fi.Path)
		}
		os.Remove(fi.Path)
	}

	log.Info("Exit")
}
