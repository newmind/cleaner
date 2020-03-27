package vods

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Day struct {
	d     int
	hours []bool
}
type Month struct {
	m    int
	days []*Day
}
type Year struct {
	y      int
	months []*Month
}

type VodInfo struct {
	path  string
	id    string
	intId int

	years []*Year
}

func (p *VodInfo) String() string {
	return p.id
}

func NewVodInfo(path string) *VodInfo {
	return &VodInfo{
		path:  path,
		id:    filepath.Base(path),
		intId: atoi(strings.Split(filepath.Base(path), "-")[0], 0),
		years: []*Year{},
	}
}

func (p *VodInfo) FillTree() {
	p.fillYear()
	p.fillMonth()
	p.fillDay()
}

func (p *VodInfo) fillYear() {
	matches, err := filepath.Glob(filepath.Join(p.path, "*"))
	if err != nil {
		log.Error(err)
		return
	}

	for _, e := range matches {
		if !isDir(e) {
			continue
		}
		year := atoi(filepath.Base(e), -1)
		if year != -1 {
			p.years = append(p.years, &Year{
				y:      year,
				months: []*Month{},
			})
		}
	}

	sort.Slice(p.years, func(i, j int) bool {
		return p.years[i].y < p.years[j].y
	})
}

func (p *VodInfo) fillMonth() {
	for _, year := range p.years {

		pattern := filepath.Join(p.path, strconv.Itoa(year.y), "*")

		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Error(err)
			return
		}

		for _, e := range matches {
			if !isDir(e) {
				continue
			}
			month := atoi(filepath.Base(e), -1)
			if month != -1 {
				year.months = append(year.months, &Month{
					m:    month,
					days: []*Day{},
				})
			}
		}

		sort.Slice(year.months, func(i, j int) bool {
			return year.months[i].m < year.months[j].m
		})
	}
}

func (p *VodInfo) fillDay() {
	for _, year := range p.years {
		for _, month := range year.months {

			pattern := filepath.Join(p.path,
				strconv.Itoa(year.y), strconv.Itoa(month.m), "*")

			matches, err := filepath.Glob(pattern)
			if err != nil {
				log.Error(err)
				return
			}

			for _, e := range matches {
				if !isDir(e) {
					continue
				}
				day := atoi(filepath.Base(e), -1)
				if day != -1 {
					month.days = append(month.days, &Day{
						d:     day,
						hours: nil,
					})
				}
			}

			sort.Slice(month.days, func(i, j int) bool {
				return month.days[i].d < month.days[j].d
			})
		}
	}
}

func (p *VodInfo) GetYears() []*Year {
	return p.years
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

func ListVODs(root string) (list []*VodInfo) {
	matches, err := filepath.Glob(filepath.Join(root, "*-0-0"))
	if err != nil {
		log.Error(err)
		return
	}
	for _, e := range matches {
		if !isDir(e) {
			continue
		}
		vodInfo := NewVodInfo(e)
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
