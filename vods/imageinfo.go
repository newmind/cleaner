package vods

import (
	"os"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

type ImageInfo struct {
	path string
	list []imageItem // add 이후에 정렬해야함(sorted by date DESC)

	modifiedForSorting bool
}

type imageItem struct {
	filename string
	modTime  time.Time
	y, m, d  int
}

func NewImageInfo(root string) *ImageInfo {
	return &ImageInfo{
		path: root,
	}
}

func (p *ImageInfo) String() string {
	return p.path
}

func (p *ImageInfo) FillTree() {
	panic("implement me")
}

func (p *ImageInfo) GetOldestDay() (found bool, year int, month int, day int) {
	if len(p.list) <= 0 {
		return
	}

	p.SortByDateDesc()

	oldestTime := p.list[len(p.list)-1].modTime
	found = true
	year = oldestTime.Year()
	month = int(oldestTime.Month())
	day = oldestTime.Day()
	return
}

func (p *ImageInfo) DeleteOldestDay(deleteLocalDir bool) {
	p.SortByDateDesc()
	found, year, month, day := p.GetOldestDay()
	if !found {
		return
	}

	for i := len(p.list) - 1; i >= 0; i-- {
		item := p.list[i]

		if year == item.y && month == item.m && day == item.d {
			log.Debugf("Delete old image [%d_%d_%d] %s", year, month, day, item.filename)
			p.list = p.list[:i]

			if deleteLocalDir {
				err := os.Remove(item.filename)
				if err != nil {
					log.Error("Failed to delete old image file : ", err)
				}
			}
		} else {
			break
		}
	}
}

func (p *ImageInfo) add(filename string, modTime time.Time) {
	p.list = append(p.list, imageItem{
		filename: filename,
		modTime:  modTime,
		y:        modTime.Year(),
		m:        int(modTime.Month()),
		d:        modTime.Day(),
	})
	p.modifiedForSorting = true
}

func (p *ImageInfo) SortByDateDesc() {
	if p.modifiedForSorting {
		sort.Slice(p.list, func(i, j int) bool {
			return p.list[i].modTime.After(p.list[j].modTime)
		})
		p.modifiedForSorting = false
	}
}
