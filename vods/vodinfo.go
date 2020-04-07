package vods

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.markany.com/argos/cleaner/common"
)

type Day struct {
	d     int
	hours []bool

	dirname string
}
type Month struct {
	m    int
	days []*Day

	dirname string
}

type Year struct {
	y      int
	months []*Month

	dirname string
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
	utc   bool
}

func (p *VodInfo) String() string {
	return p.id
}

func (p *VodInfo) Id() int {
	return p.intId
}

func NewVodInfo(root string, id string, isUTC bool) *VodInfo {
	return &VodInfo{
		path:  filepath.Join(root, id),
		id:    id,
		intId: common.Atoi(strings.Split(id, "-")[0], 0),
		years: []*Year{},
		utc:   isUTC,
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
		if !common.IsDir(e) {
			continue
		}
		year := common.Atoi(filepath.Base(e), -1)
		if year != -1 {
			p.years = append(p.years, &Year{
				y:       year,
				months:  []*Month{},
				dirname: filepath.Base(e),
			})
		}
	}

	sort.Slice(p.years, func(i, j int) bool {
		return p.years[i].y < p.years[j].y
	})
}

func (p *VodInfo) fillMonth() {
	for _, year := range p.years {

		pattern := filepath.Join(p.path, year.dirname, "*")

		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Error(err)
			return
		}

		for _, e := range matches {
			if !common.IsDir(e) {
				continue
			}
			month := common.Atoi(filepath.Base(e), -1)
			if month != -1 {
				year.months = append(year.months, &Month{
					m:       month,
					days:    []*Day{},
					dirname: filepath.Base(e),
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

			pattern := filepath.Join(p.path, year.dirname, month.dirname, "*")

			matches, err := filepath.Glob(pattern)
			if err != nil {
				log.Error(err)
				return
			}

			for _, e := range matches {
				if !common.IsDir(e) {
					continue
				}
				day := common.Atoi(filepath.Base(e), -1)
				if day != -1 {
					month.days = append(month.days, &Day{
						d:       day,
						hours:   nil,
						dirname: filepath.Base(e),
					})
				}
			}

			sort.Slice(month.days, func(i, j int) bool {
				return month.days[i].d < month.days[j].d
			})
		}
	}
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

func (p *VodInfo) GetOldestDateUTC() (found bool, dateUTC time.Time) {
	found, year, month, day := p.GetOldestDay()
	if !found {
		return
	}
	if p.utc {
		dateUTC = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	} else {
		dateUTC = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).UTC()
	}
	return
}

func (p *VodInfo) DeleteOldestDay(deleteLocalDir bool) {
	var found bool
	var yIdx, mIdx, dIdx int

	log := log.WithFields(log.Fields{
		"id":   p.id,
		"root": p.path,
	})

LOOP:
	for i, yy := range p.years {
		for j, mm := range yy.months {
			for k, _ := range mm.days {
				found = true
				yIdx, mIdx, dIdx = i, j, k
				break LOOP
			}
		}
	}

	if found {
		// delete day
		year := p.years[yIdx]
		month := year.months[mIdx]
		day := month.days[dIdx]

		log.Debugf("Delete dir [%s] %v/%v/%v %s(utc=%v)", p.id, year.dirname, month.dirname, day.dirname, p.path, p.utc)

		month.deleteDayByIndex(dIdx)
		if deleteLocalDir {
			toDelete := filepath.Join(p.path, year.dirname, month.dirname, day.dirname)
			// Windows 에서는 삭제가 바로 안되는 문제 있음.
			// https://github.com/golang/go/issues/20841
			// 그래서 RemoveAll 이후 상위 디렉토리를 os.Remove 로 삭제시 실패남
			err := os.RemoveAll(toDelete)
			if err != nil {
				log.Error(err)
			}
		}

		if month.isEmpty() {
			year.deleteMonthByIndex(mIdx)
			if deleteLocalDir {
				toDelete := filepath.Join(p.path, year.dirname, month.dirname)
				err := os.Remove(toDelete)
				if err != nil {
					log.Error(err)
				}
			}
		}
	} else {
		log.Warnln("DeleteOldestDay : Not found")
	}
}

func (p *VodInfo) add(year int, month int, day int) {
	var (
		yP *Year
		mP *Month
		dP *Day
	)

	for _, y := range p.years {
		if y.y == year {
			yP = y
			break
		}
	}
	if yP == nil {
		yP = &Year{
			y:      year,
			months: []*Month{},
		}
	}

	for _, m := range yP.months {
		if m.m == month {
			mP = m
			break
		}
	}
	if mP == nil {
		mP = &Month{
			m:    month,
			days: []*Day{},
		}
	}

	for _, d := range mP.days {
		if d.d == day {
			dP = d
			break
		}
	}
	if dP == nil {
		dP = &Day{
			d:     day,
			hours: nil,
		}
	}

	//
	mP.days = append(mP.days, dP)
	sort.Slice(mP.days, func(i, j int) bool {
		return mP.days[i].d < mP.days[j].d
	})

	yP.months = append(yP.months, mP)
	sort.Slice(yP.months, func(i, j int) bool {
		return yP.months[i].m < yP.months[j].m
	})

	p.years = append(p.years, yP)
	sort.Slice(p.years, func(i, j int) bool {
		return p.years[i].y < p.years[j].y
	})
}
