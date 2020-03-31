package watcher

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.markany.com/argos/cleaner/fileinfo"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const root = "../test"

//const root string = "/Volumes/RAMDisk"

func init() {
	os.Mkdir(root, os.ModePerm)
	log.SetLevel(log.TraceLevel)
}

func TestWatch(t *testing.T) {
	ch := make(chan fileinfo.FileInfo, 100)

	done := make(chan bool)

	var addedFile string

	go func() {
		select {
		case fi := <-ch:
			t.Log(fi)
			// fmt.Println(fi)
			assert.Equal(t, addedFile, fi.Path)
			defer func() {
				// time.Sleep(1 * time.Millisecond)
				done <- true
			}()
		}
	}()

	err := Watch(root, ch)
	if err != nil {
		t.Fatal(err)
	}

	tmpdir := filepath.Join(root, "t")
	os.Mkdir(tmpdir, os.ModePerm)
	defer os.Remove(tmpdir)

	//sleep 없이 하위 디렉토리에 대한 이벤트 감지 할수 있어야 함
	time.Sleep(1 * time.Millisecond)

	// test to create a file
	f, err := ioutil.TempFile(tmpdir, "test*")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(f.Name())

	addedFile = f.Name()

	<-done
}

//func TestExcludedDir(t *testing.T) {
//	excludedDir := filepath.Join(root, "excluded")
//
//	// add to Excludes
//	excludes.Add(excludedDir)
//	defer func() {
//		excludes.Remove(excludedDir)
//	}()
//
//	ch := make(chan fileinfo.FileInfo, 100)
//
//	done := make(chan bool)
//
//	var addedFile string
//	var tmpExcludedFile string
//	var deletedFile string
//
//	go func() {
//		select {
//		case fi := <-ch:
//			t.Log(fi)
//			//fmt.Println(fi)
//			deletedFile = fi.Path
//			assert.NotEqual(t, tmpExcludedFile, fi.Path)
//			assert.Equal(t, addedFile, fi.Path)
//			// defer func() {
//			// time.Sleep(1 * time.Millisecond)
//			done <- true
//			// }()
//		}
//	}()
//
//	err := Watch(root, ch)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// 1. Creates a file in excluded directory
//	os.Mkdir(excludedDir, os.ModePerm)
//	time.Sleep(time.Millisecond * 10)
//	defer os.Remove(excludedDir)
//
//	// 디렉토리 생성후, 잠시 쉬어야 다음 파일 생성 이벤트 받을수 있음
//	time.Sleep(1 * time.Millisecond)
//
//	// should skip this file
//	tmpExcludedFile = filepath.Join(excludedDir, "somefile")
//	if err := ioutil.WriteFile(tmpExcludedFile, []byte("Hello"), os.ModePerm); err != nil {
//		t.Fatal(err)
//	}
//	defer os.Remove(tmpExcludedFile)
//
//	// 2. Creates another non-excluded file
//	f, err := ioutil.TempFile(root, "test*")
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	defer os.Remove(f.Name())
//
//	addedFile = f.Name()
//
//	<-done
//	t.Log(addedFile, deletedFile)
//	assert.Equal(t, addedFile, deletedFile)
//}
//
//func TestMkdirAll(t *testing.T) {
//	ch := make(chan fileinfo.FileInfo, 100)
//
//	done := make(chan bool)
//
//	var addedFile string
//
//	go func() {
//		select {
//		case fi := <-ch:
//			t.Log(fi)
//			// fmt.Println(fi)
//			assert.Equal(t, addedFile, fi.Path)
//			defer func() {
//				// time.Sleep(1 * time.Millisecond)
//				done <- true
//			}()
//		}
//	}()
//
//	err := Watch(root, ch)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	tmpdir := filepath.Join(root, "t", "t2", "t3")
//	os.MkdirAll(tmpdir, os.ModePerm)
//	time.Sleep(time.Millisecond * 10)
//	defer os.RemoveAll(filepath.Join(root, "t"))
//
//	// test to create a file
//	f, err := ioutil.TempFile(tmpdir, "test*")
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	defer os.Remove(f.Name())
//
//	addedFile = f.Name()
//
//	<-done
//}
