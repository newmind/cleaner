package scanner

import (
	"os"
	"path/filepath"

	"gitlab.markany.com/argos/cleaner/fileinfo"

	"github.com/karrick/godirwalk"
	"github.com/spf13/viper"
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
	deleteHidden := viper.GetBool("delete_hidden")
	files = make([]*fileinfo.FileInfo, 0, 1000)

	err = godirwalk.Walk(root, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			base := filepath.Base(path)
			if !de.IsDir() {
				if base[0] == '.' && deleteHidden == false {
					return nil
				}
				fi, err := os.Lstat(path)
				if err == nil {
					files = append(files,
						&fileinfo.FileInfo{
							Path: path,
							Time: fi.ModTime(),
						})
				}
			} else if de.IsDir() {
				if base[0] == '.' && deleteHidden == false {
					return filepath.SkipDir
				}
			}
			return nil
		},

		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})

	return
}
