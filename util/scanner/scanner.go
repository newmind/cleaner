package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gitlab.markany.com/argos/cleaner/scanner"
)

const appName = "scanner"

func init() {
	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Infof("Starting %v...\n", appName)

	path := flag.String("path", "", "[Required] Path where to create files")
	flag.Parse()

	files := scanner.ScanAllFiles([]string{*path})
	log.Infof("Scanned %d files", len(files))
}
