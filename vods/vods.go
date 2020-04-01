package vods

import (
	"os"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"
	"gitlab.markany.com/argos/cleaner/common"
)

func ListAllVODs(root string) (list []*VodInfo) {
	list = []*VodInfo{}
	if _, err := os.Stat(root); err != nil && os.IsNotExist(err) {
		return
	}

	matches, err := filepath.Glob(filepath.Join(root, "*-0-0"))
	if err != nil {
		log.Error(err)
		return
	}
	for _, e := range matches {
		if !common.IsDir(e) {
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

//ListOldestCCTV : list 내에서 가장 오래된 날짜의 모든 목록 리턴
func ListOldestCCTV(list []*VodInfo) (result []*VodInfo) {
	result = []*VodInfo{}
	var (
		minY = 9999
		minM = 12 + 1
		minD = 31 + 1
	)
	for _, cctv := range list {
		found, y, m, d := cctv.GetOldestDay()
		if found {
			if y < minY ||
				y <= minY && m < minM ||
				y <= minY && m <= minM && d < minD {
				result = []*VodInfo{cctv}
				minY, minM, minD = y, m, d
			} else if y == minY && m == minM && d == minD {
				result = append(result, cctv)
			}
		}
	}
	return
}
