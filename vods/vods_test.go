package vods

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)
import "github.com/stretchr/testify/assert"

const root string = "../test/vods"

func init() {
	os.Mkdir(root, os.ModePerm)
}

type TestVODInfo struct {
	path, id string
	Years    map[int]map[int][]int `json:"years"`
}

func (p *TestVODInfo) deleteLocal() {
	os.RemoveAll(p.path)
}

func generateTestVOD(root string, id string, jsonVodTree string) (error, *TestVODInfo) {
	tree := &TestVODInfo{
		path: filepath.Join(root, id),
		id:   id,
	}
	err := json.Unmarshal([]byte(jsonVodTree), tree)
	if err != nil {
		return err, nil
	}

	os.Mkdir(filepath.Join(root, id), os.ModePerm)
	for year, months := range tree.Years {
		os.Mkdir(filepath.Join(root, id, strconv.Itoa(year)), os.ModePerm)
		for month, days := range months {
			os.Mkdir(filepath.Join(root, id, strconv.Itoa(year), strconv.Itoa(month)), os.ModePerm)
			for _, day := range days {
				os.Mkdir(filepath.Join(root, id, strconv.Itoa(year), strconv.Itoa(month), strconv.Itoa(day)), os.ModePerm)
			}
		}
	}

	return nil, tree
}

func TestListAllVODs(t *testing.T) {
	_, v1 := generateTestVOD(root, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	_, v2 := generateTestVOD(root, "776-0-0", `{"years": 
        {"2020":{ "2":[17,18,19],
				  "3":[1,2,3,4,5,6,7,8,9]
			    }
		}
	}`)
	if v2 != nil {
		defer v2.deleteLocal()
	}

	_, v3 := generateTestVOD(root, "2-0-0", `{"years": {
		"2020":{ "1":[17,18,19],
				 "2":[1,2,3,4,5,6,7,8,9]
			   }
		}
	}`)
	if v3 != nil {
		defer v3.deleteLocal()
	}

	list := ListAllVODs(root)
	assert.Equal(t, 3, len(list))
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
	_, v1 := generateTestVOD(root, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(root)
	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)
}

func TestDeleteOldestVOD(t *testing.T) {
	_, v1 := generateTestVOD(root, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(root)

	// 하루치 삭제
	list[0].DeleteOldestDay(true)

	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 14, d)

	// 1d월 14일부터 ~ 31일까지 삭제
	for d := 14; d <= 31; d++ {
		list[0].DeleteOldestDay(true)
	}

	// reload list from disk
	list = ListAllVODs(root)

	// 2월 데이터 나와야 함
	found, y, m, d = list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)
}

func TestEmptyDir(t *testing.T) {
	_, v1 := generateTestVOD(root, "1-0-0", `{"years": 
		{"2020":{ "1":[],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(root)

	// 비어있는 폴더는 제외하는지 체크
	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)

	// 비어있는 달,일 체크
	_, v1 = generateTestVOD(root, "1-0-0", `{"years": 
		{
         "2018":{},
		 "2019":{ "11":[],
				  "12":[]
		 	    },
		 "2020":{ "1":[],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}
	found, y, m, d = list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)
}

func TestMonthChanged(t *testing.T) {
	_, v1 := generateTestVOD(root, "1-0-0", `{"years": 
		{"2020":{ "1":[31],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(root)

	// 하루치 삭제
	list[0].DeleteOldestDay(true)

	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)
}

func TestYearChanged(t *testing.T) {
	_, v1 := generateTestVOD(root, "1-0-0", `{"years": 
		{
		 "2019":{ "11":[],
				  "12":[31]
		 	    },
		 "2020":{ "1":[13],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(root)

	// 하루치 삭제
	list[0].DeleteOldestDay(true)

	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)

	// reload
	list = ListAllVODs(root)
	found, y, m, d = list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)
}

func TestListOldestCCTV(t *testing.T) {
	list := ListAllVODs(root)
	oldestCCTVs := ListOldestCCTV(list)

	for _, v := range oldestCCTVs {
		found, y, m, d := v.GetOldestDay()
		t.Log(v.id, found, y, m, d)
	}
}
