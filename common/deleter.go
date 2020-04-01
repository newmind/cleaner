package common

type IDeleteOld interface {
	String() string
	FillTree()
	GetOldestDay() (found bool, year int, month int, day int)
	DeleteOldestDay(deleteLocalDir bool)
}
