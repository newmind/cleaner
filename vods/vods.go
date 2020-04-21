package vods

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.markany.com/argos/cleaner/common"
)

func ListAllVODs(root string) (list []ICommonDeleter) {
	list = []ICommonDeleter{}
	if _, err := os.Stat(root); err != nil && os.IsNotExist(err) {
		log.Warn("vods directory doesn't exist : ", root)
		return
	}

	matches, err := filepath.Glob(filepath.Join(root, "*-0-0", "UTC"))
	if err == nil {
		for _, e := range matches {
			if !common.IsDir(e) {
				continue
			}
			id := filepath.Base(filepath.Dir(e))
			vodInfo := NewVodInfo(filepath.Dir(e), id, true, e)
			vodInfo.FillTree()
			//log.Debugf("%#v", vodInfo)

			list = append(list, vodInfo)
		}
	}

	matches, err = filepath.Glob(filepath.Join(root, "*-0-0"))
	if err == nil {
		for _, e := range matches {
			if !common.IsDir(e) {
				continue
			}
			vodInfo := NewVodInfo(filepath.Dir(e), filepath.Base(e), false, e)
			vodInfo.FillTree()
			//log.Debugf("%#v", vodInfo)

			list = append(list, vodInfo)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Id() < list[j].Id()
	})
	return
}

//FilterOldestDay : list 내에서 가장 오래된 날짜의 모든 목록 리턴
func FilterOldestDay(list []ICommonDeleter) (result []ICommonDeleter) {
	result = []ICommonDeleter{}
	var (
		minY = 9999
		minM = 12 + 1
		minD = 31 + 1
	)
	for _, info := range list {
		found, y, m, d := info.GetOldestDay()
		if found {
			if y < minY ||
				y <= minY && m < minM ||
				y <= minY && m <= minM && d < minD {
				result = []ICommonDeleter{info}
				minY, minM, minD = y, m, d
			} else if y == minY && m == minM && d == minD {
				result = append(result, info)
			}
		}
	}
	return
}

func ListAllImages(root string) (list []ICommonDeleter) {
	list = []ICommonDeleter{}

	if _, err := os.Stat(root); err != nil && os.IsNotExist(err) {
		log.Warnf("%s directory does not exist : ", root)
		return
	}

	// 1. UTC 폴더/파일이름 포맷,  vods 와 동일한 폴더구조
	matches, err := filepath.Glob(filepath.Join(root, "*"))
	if err == nil {
		for _, e := range matches {
			if !common.IsDir(e) {
				continue
			}
			vodInfo := NewVodInfo(filepath.Dir(e), filepath.Base(e), true, e)
			vodInfo.FillTree()
			//log.Debugf("%#v", vodInfo)

			list = append(list, vodInfo)
		}
	}

	// 2. 이전 포맷 데이터. jpg 파일들이 /Images 폴더 한군데에 전부 있음
	matches, err = filepath.Glob(filepath.Join(root, "*.jpg"))
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

// DeleteOlderThan : retentionDays 보다 오래된 날짜를 지움. timezone 은 무시
func DeleteOlderThan(allVodList []ICommonDeleter, retentionDays int, dryRun bool) {
	nowUTC := time.Now().UTC()
	retentionDate := nowUTC.Add(-time.Hour * 24 * time.Duration(retentionDays))

	for _, vodInfo := range allVodList {
		for {
			found, dateUTC := vodInfo.GetOldestDateUTC()
			if found && dateUTC.Before(retentionDate) {
				vodInfo.DeleteOldestDay(!dryRun)
			} else {
				break
			}
		}
	}
}

func DeleteOldest(allVodList []ICommonDeleter, dryRun bool) (deleted bool) {
	var oldInfos = FilterOldestDay(allVodList)
	//log.Debug("oldInfos len = ", len(oldInfos))

	//var found bool
	if len(oldInfos) > 0 {
		found, y, m, d := oldInfos[0].GetOldestDay()
		_, _, _ = y, m, d
		if found {
			oldInfos[0].DeleteOldestDay(!dryRun)
			return true
		}
	}
	return
}
