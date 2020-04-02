package vods

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestImageInfo_DeleteOldestDay(t *testing.T) {
	imageInfo := NewImageInfo("whatever")

	found, _, _, _ := imageInfo.GetOldestDay()
	assert.False(t, found, "Check empty")

	imageInfo.AddToLast("4_2", time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC))
	imageInfo.AddToLast("4_1", time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC))

	found, year, month, day := imageInfo.GetOldestDay()
	assert.Equal(t, year, 2020)
	assert.Equal(t, month, 4)
	assert.Equal(t, day, 1)

	imageInfo.DeleteOldestDay(false)
	found, year, month, day = imageInfo.GetOldestDay()
	assert.Equal(t, year, 2020)
	assert.Equal(t, month, 4)
	assert.Equal(t, day, 2)
}
