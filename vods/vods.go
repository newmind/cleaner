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
		log.Warn("vods directory doesn't exist : ", root)
		return
	}

	matches, err := filepath.Glob(filepath.Join(root, "UTC", "*-0-0"))
	if err == nil {
		for _, e := range matches {
			if !common.IsDir(e) {
				continue
			}
			vodInfo := NewVodInfo(filepath.Dir(e), filepath.Base(e), true)
			vodInfo.FillTree()

			list = append(list, vodInfo)
		}
	}

	matches, err = filepath.Glob(filepath.Join(root, "*-0-0"))
	if err == nil {
		for _, e := range matches {
			if !common.IsDir(e) {
				continue
			}
			vodInfo := NewVodInfo(filepath.Dir(e), filepath.Base(e), false)
			vodInfo.FillTree()

			list = append(list, vodInfo)
		}
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

func ListAllImages(root string) (list []*ImageInfo) {
	list = []*ImageInfo{}

	if _, err := os.Stat(root); err != nil && os.IsNotExist(err) {
		log.Warnf("%s directory does not exist : ", root)
		return
	}

	matches, err := filepath.Glob(filepath.Join(root, "*.jpg"))
	if err != nil {
		log.Error(err)
		return
	}

	imageInfo := NewImageInfo(root)
	list = append(list, imageInfo)

	for _, e := range matches {
		if common.IsDir(e) {
			continue
		}
		info, err := os.Stat(e)
		if os.IsNotExist(err) {
			continue
		}
		imageInfo.add(e, info.ModTime())
	}

	return
}
