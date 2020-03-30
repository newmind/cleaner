package vods

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func ListAllVODs(root string) (list []*VodInfo) {
	matches, err := filepath.Glob(filepath.Join(root, "*-0-0"))
	if err != nil {
		log.Error(err)
		return
	}
	for _, e := range matches {
		if !isDir(e) {
			continue
		}
		vodInfo := NewVodInfo(filepath.Dir(e), filepath.Base(e))
		vodInfo.FillTree()

		list = append(list, vodInfo)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].intId < list[j].intId
	})
	return
}

func isDir(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	if ret, err := strconv.Atoi(s); err == nil {
		return ret
	}
	return def
}

func ListOldestCCTV(list []*VodInfo) (result []*VodInfo) {
	var (
		minY = 9999
		minM = 12 + 1
		minD = 31 + 1
	)
	for _, cctv := range list {
		found, y, m, d := cctv.GetOldestDay()
		if found {
			if y < minY && m < minM && d < minD {
				result = []*VodInfo{cctv}
				minY, minM, minD = y, m, d
			} else if y == minY && m == minM && d == minD {
				result = append(result, cctv)
			}
		}
	}
	return
}
