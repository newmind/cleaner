package scanner

import (
	"os"

	"github.com/karrick/godirwalk"
	"gitlab.markany.com/argos/cleaner/fileinfo"
)

func GoDirWalk(root string) ([]string, error) {
	var files []string
	err := godirwalk.Walk(root, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				files = append(files, path)
			}
			// fmt.Printf("%s %s\n", de.ModeType(), path)
			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})
	return files, err
}

// GoFileWalk root 디렉토리내의 모든 파일을 리턴
func GoFileWalk(root string) (files []*fileinfo.FileInfo, err error) {
	files = make([]*fileinfo.FileInfo, 0, 1000)

	err = godirwalk.Walk(root, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				fi, err := os.Lstat(path)
				if err == nil {
					files = append(files,
						&fileinfo.FileInfo{
							Path: path,
							Time: fi.ModTime(),
						})
				}
			}
			return nil
		},

		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})

	return
}
