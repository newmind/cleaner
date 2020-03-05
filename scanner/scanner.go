package scanner

import (
	"sort"

	"github.com/sirupsen/logrus"
	"gitlab.markany.com/argos/cleaner/fileinfo"
)

// ScanAllFiles scans all filess in dirs, and sort by date DESC
func ScanAllFiles(dirs []string) []*fileinfo.FileInfo {
	var scannedFiles []*fileinfo.FileInfo
	for _, dir := range dirs {
		logrus.Infof("Scanning directory %v ...", dir)
		files, err := GoFileWalk(dir)
		if err == nil {
			scannedFiles = append(scannedFiles, files...)
		}
	}
	// 날짜순 ** DESC 정렬 **
	sort.Slice(scannedFiles, func(i, j int) bool {
		return scannedFiles[i].Time.UnixNano() > scannedFiles[j].Time.UnixNano()
	})
	return scannedFiles
}
