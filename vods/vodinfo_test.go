package vods

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVodInfo_DeleteOldestDay(t *testing.T) {
	vodInfo := NewVodInfo("dummy", "1", false)

	found, _, _, _ := vodInfo.GetOldestDay()
	assert.False(t, found, "Check empty")

	vodInfo.add(2020, 4, 2)
	found, year, month, day := vodInfo.GetOldestDay()
	assert.Equal(t, year, 2020)
	assert.Equal(t, month, 4)
	assert.Equal(t, day, 2)

	vodInfo.add(2020, 4, 1)
	found, year, month, day = vodInfo.GetOldestDay()
	assert.Equal(t, year, 2020)
	assert.Equal(t, month, 4)
	assert.Equal(t, day, 1)

	vodInfo.add(2020, 3, 20)
	found, year, month, day = vodInfo.GetOldestDay()
	assert.Equal(t, year, 2020)
	assert.Equal(t, month, 3)
	assert.Equal(t, day, 20)

	vodInfo.add(2010, 11, 25)
	found, year, month, day = vodInfo.GetOldestDay()
	assert.Equal(t, year, 2010)
	assert.Equal(t, month, 11)
	assert.Equal(t, day, 25)

	vodInfo.DeleteOldestDay(false)
	found, year, month, day = vodInfo.GetOldestDay()
	assert.Equal(t, year, 2020)
	assert.Equal(t, month, 3)
	assert.Equal(t, day, 20)
}
