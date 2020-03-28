package vods

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

const root string = "../test/vods"

func TestListAllVODs(t *testing.T) {
	list := ListAllVODs(root)
	assert.Equal(t, 122, len(list))
	t.Log(list)
	//
	firstCam := list[0]
	assert.Equal(t, "1-0-0", firstCam.id)
	assert.Equal(t, 2020, firstCam.years[0].y)
	assert.Equal(t, 1, firstCam.years[0].months[0].m)
	assert.Equal(t, 13, firstCam.years[0].months[0].days[0].d)

	lastCam := list[len(list)-1]
	assert.Equal(t, "776-0-0", lastCam.id)
	assert.Equal(t, 2020, lastCam.years[0].y)
	assert.Equal(t, 2, lastCam.years[0].months[0].m)
	assert.Equal(t, 17, lastCam.years[0].months[0].days[0].d)
}

func TestGetOldestVOD(t *testing.T) {
	list := ListAllVODs(root)
	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)
}

func TestDeleteOldestVOD(t *testing.T) {
	list := ListAllVODs(root)

	list[0].DeleteOldestDay(false)

	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 14, d)

	for d := 14; d <= 31; d++ {
		list[0].DeleteOldestDay(false)
	}

	found, y, m, d = list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)
}

func TestListOldestCCTV(t *testing.T) {
	list := ListAllVODs(root)
	oldestCCTVs := ListOldestCCTV(list)

	for _, v := range oldestCCTVs {
		found, y, m, d := v.GetOldestDay()
		t.Log(v.id, found, y, m, d)
	}
}
