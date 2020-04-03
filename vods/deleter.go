package vods

import "time"

type ICommonDeleter interface {
	String() string
	FillTree()
	GetOldestDay() (found bool, year int, month int, day int)
	GetOldestDateUTC() (found bool, dateUTC time.Time)
	DeleteOldestDay(deleteLocalDir bool)
	Id() int
}
