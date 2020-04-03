package vods

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)
import "github.com/stretchr/testify/assert"

const rootVod string = "../test/vods"
const rootImage string = "../test/images"

func init() {
	_ = os.Mkdir(rootVod, os.ModePerm)
}

type TestVODInfo struct {
	path, id string
	Years    map[int]map[int][]int `json:"years"`
}

func (p *TestVODInfo) deleteLocal() {
	_ = os.RemoveAll(p.path)
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

	_ = os.Mkdir(filepath.Join(root, id), os.ModePerm)
	for year, months := range tree.Years {
		_ = os.Mkdir(filepath.Join(root, id, strconv.Itoa(year)), os.ModePerm)
		for month, days := range months {
			_ = os.Mkdir(filepath.Join(root, id, strconv.Itoa(year), strconv.Itoa(month)), os.ModePerm)
			for _, day := range days {
				_ = os.Mkdir(filepath.Join(root, id, strconv.Itoa(year), strconv.Itoa(month), strconv.Itoa(day)), os.ModePerm)
			}
		}
	}

	return nil, tree
}

func TestListAllVODs(t *testing.T) {
	os.RemoveAll(rootVod)
	time.Sleep(time.Millisecond * 10)
	os.Mkdir(rootVod, os.ModePerm)
	time.Sleep(time.Millisecond * 10)

	_, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	_, v2 := generateTestVOD(rootVod, "776-0-0", `{"years": 
        {"2020":{ "2":[17,18,19],
				  "3":[1,2,3,4,5,6,7,8,9]
			    }
		}
	}`)
	if v2 != nil {
		defer v2.deleteLocal()
	}

	_, v3 := generateTestVOD(rootVod, "2-0-0", `{"years": {
		"2020":{ "1":[17,18,19],
				 "2":[1,2,3,4,5,6,7,8,9]
			   }
		}
	}`)
	if v3 != nil {
		defer v3.deleteLocal()
	}

	list := ListAllVODs(rootVod)
	assert.Equal(t, 3, len(list))
	t.Log(list)
	//
	firstCam := list[0]
	_, year, month, day := firstCam.GetOldestDay()
	assert.Equal(t, "1-0-0", firstCam.String())
	assert.Equal(t, 2020, year)
	assert.Equal(t, 1, month)
	assert.Equal(t, 13, day)

	lastCam := list[len(list)-1]
	_, year, month, day = lastCam.GetOldestDay()
	assert.Equal(t, "776-0-0", lastCam.String())
	assert.Equal(t, 2020, year)
	assert.Equal(t, 2, month)
	assert.Equal(t, 17, day)
}

func TestGetOldestVOD(t *testing.T) {
	os.RemoveAll(rootVod)
	os.Mkdir(rootVod, os.ModePerm)

	_, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(rootVod)
	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)
}

func TestDeleteOldestVOD(t *testing.T) {
	os.RemoveAll(rootVod)
	os.Mkdir(rootVod, os.ModePerm)

	_, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(rootVod)

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

	// 2월 데이터 나와야 함
	found, y, m, d = list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)

	// reload list from disk
	list2 := ListAllVODs(rootVod)
	assert.Equal(t, list2, list, "should be equal, after reloading")
}

func TestEmptyRoot(t *testing.T) {
	os.RemoveAll(rootVod)
	list := ListAllVODs(rootVod)
	assert.Len(t, list, 0)
	assert.NotNil(t, list)

	_ = os.Mkdir(rootVod, os.ModePerm)
	list = ListAllVODs(rootVod)
	assert.Len(t, list, 0)
	assert.NotNil(t, list)

	oldestCCTV := FilterOldestDay(list)
	assert.Len(t, oldestCCTV, 0)
	assert.NotNil(t, oldestCCTV)

	// 비어있는 달,일 체크
	err, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{
         "2018":{},
		 "2020":{ "1":[]
		 	    }
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if v1 != nil {
		defer v1.deleteLocal()
	}
	list = ListAllVODs(rootVod)
	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.False(t, found)
}

func TestEmptyDir(t *testing.T) {
	os.RemoveAll(rootVod)
	os.Mkdir(rootVod, os.ModePerm)

	_, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{"2020":{ "1":[],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(rootVod)

	// 비어있는 폴더는 제외하는지 체크
	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)

	// 비어있는 달,일 체크
	_, v1 = generateTestVOD(rootVod, "1-0-0", `{"years": 
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
	os.RemoveAll(rootVod)
	os.Mkdir(rootVod, os.ModePerm)

	_, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{"2020":{ "1":[31],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if v1 != nil {
		defer v1.deleteLocal()
	}

	list := ListAllVODs(rootVod)

	// 하루치 삭제
	list[0].DeleteOldestDay(true)

	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 2, m)
	assert.Equal(t, 1, d)
}

func TestYearChanged(t *testing.T) {
	os.RemoveAll(rootVod)
	os.Mkdir(rootVod, os.ModePerm)

	_, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
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

	list := ListAllVODs(rootVod)

	// 하루치 삭제
	list[0].DeleteOldestDay(true)

	found, y, m, d := list[0].GetOldestDay()
	t.Log(found, y, m, d)
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)

	// reload
	list2 := ListAllVODs(rootVod)
	assert.Equal(t, list2, list, "should be equal, after reloading")
}

func TestFilterOldestDay(t *testing.T) {
	os.RemoveAll(rootVod)
	os.Mkdir(rootVod, os.ModePerm)

	// 순차적으로 여러버 테스트
	err, v1 := generateTestVOD(rootVod, "1-0-0", `{"years": 
		{"2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if v1 != nil {
		defer v1.deleteLocal()
	}

	err, v2 := generateTestVOD(rootVod, "2-0-0", `{"years": 
		{"2019":{ "12":[15]
                },
		 "2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
				}
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if v2 != nil {
		defer v2.deleteLocal()
	}

	err, v3 := generateTestVOD(rootVod, "3-0-0", `{"years": 
		{"2020":{ "1":[13,14,15],
				  "2":[1,2,3,4,5,6,7,8,9]
		 	    }
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if v3 != nil {
		defer v3.deleteLocal()
	}

	list := ListAllVODs(rootVod)

	// get
	oldestCCTVs := FilterOldestDay(list)
	assert.Equal(t, 1, len(oldestCCTVs), "length == 1")
	_, y, m, d := oldestCCTVs[0].GetOldestDay()
	assert.Equal(t, 2019, y)
	assert.Equal(t, 12, m)
	assert.Equal(t, 15, d)

	// delete
	oldestCCTVs[0].DeleteOldestDay(true)

	// get
	oldestCCTVs = FilterOldestDay(list)
	assert.Equal(t, 3, len(oldestCCTVs), "length == 3")
	_, y, m, d = oldestCCTVs[0].GetOldestDay()
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)

	// delete
	oldestCCTVs[0].DeleteOldestDay(true)
	oldestCCTVs = FilterOldestDay(list)
	assert.Equal(t, 2, len(oldestCCTVs), "length == 2")
	assert.Equal(t, 2020, y)
	assert.Equal(t, 1, m)
	assert.Equal(t, 13, d)

	// after reloading
	list2 := ListAllVODs(rootVod)
	assert.Equal(t, list2, list, "should be equal after reload")
}
