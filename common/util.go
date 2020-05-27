package common

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func Atoi(s string, def int) int {
	if s == "" {
		return def
	}
	if ret, err := strconv.Atoi(s); err == nil {
		return ret
	}
	return def
}

// dir 안의 파일을 개별적으로 삭제한후 폴더 삭제. 성능 영향
func RemoveAll(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fileInfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return err
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime().Unix() < fileInfos[j].ModTime().Unix()
	})

	for _, file := range fileInfos {
		err = os.Remove(filepath.Join(path, file.Name()))
		if err != nil {
			log.Error(err)
		}
		time.Sleep(0) // 속도 조절
	}

	return os.RemoveAll(path)
}
