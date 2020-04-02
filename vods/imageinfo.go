package vods

import (
	"os"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

type ImageInfo struct {
	path string
	list []imageItem // sorted by date DESC
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

	oldestTime := p.list[len(p.list)-1].modTime
	found = true
	year = oldestTime.Year()
	month = int(oldestTime.Month())
	day = oldestTime.Day()
	return
}

func (p *ImageInfo) DeleteOldestDay(deleteLocalDir bool) {
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

func (p *ImageInfo) Add(filename string, modTime time.Time) {
	p.list = append(p.list, imageItem{
		filename: filename,
		modTime:  modTime,
		y:        modTime.Year(),
		m:        int(modTime.Month()),
		d:        modTime.Day(),
	})
}

func (p *ImageInfo) SortByDateDesc() {
	sort.Slice(p.list, func(i, j int) bool {
		return p.list[i].modTime.After(p.list[j].modTime)
	})
}
