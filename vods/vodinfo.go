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

func (y *Year) deleteMonthByIndex(monthIndex int) {
	y.months = append(y.months[:monthIndex], y.months[monthIndex+1:]...)
}

func (m *Month) deleteDayByIndex(dayIndex int) {
	m.days = append(m.days[:dayIndex], m.days[dayIndex+1:]...)
}

func (m *Month) isEmpty() bool {
	return m.days == nil || len(m.days) == 0
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

func NewVodInfo(root string, id string) *VodInfo {
	return &VodInfo{
		path:  filepath.Join(root, id),
		id:    id,
		intId: atoi(strings.Split(id, "-")[0], 0),
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

func (p *VodInfo) GetOldestDay() (found bool, year int, month int, day int) {
	for _, y := range p.years {
		for _, m := range y.months {
			for _, d := range m.days {
				return true, y.y, m.m, d.d
			}
		}
	}
	return
}

func (p *VodInfo) DeleteOldestDay(deleteLocalDir bool) {
	var found bool
	var yIdx, mIdx, dIdx int
	var y, m, d int

LOOP:
	for i, yy := range p.years {
		for j, mm := range yy.months {
			for k, dd := range mm.days {
				found = true
				yIdx, mIdx, dIdx = i, j, k
				y, m, d = yy.y, mm.m, dd.d
				break LOOP
			}
		}
	}

	if found {
		log.Debugf("DeleteOldestDay [%s] %d-%d-%d\n", p.id, y, m, d)
		// delete day
		month := p.years[yIdx].months[mIdx]
		month.deleteDayByIndex(dIdx)
		if deleteLocalDir {
			toDelete := filepath.Join(p.path, strconv.Itoa(y), strconv.Itoa(m), strconv.Itoa(d))
			err := os.RemoveAll(toDelete)
			if err != nil {
				log.Error(err)
			}
		}

		if month.isEmpty() {
			p.years[yIdx].deleteMonthByIndex(mIdx)
			if deleteLocalDir {
				toDelete := filepath.Join(p.path, strconv.Itoa(y), strconv.Itoa(m))
				err := os.Remove(toDelete)
				if err != nil {
					log.Error(err)
				}
			}
		}
	} else {
		log.Debugln("DeleteOldestDay : Not found")
	}
}
