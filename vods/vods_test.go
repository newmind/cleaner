package vods

import "testing"
import "github.com/stretchr/testify/assert"

const root string = "../test/vods"

func TestListCCTV(t *testing.T) {
	list := ListVODs(root)
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
