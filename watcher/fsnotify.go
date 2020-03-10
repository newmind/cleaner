package watcher

import (
	"container/list"
	"os"
	"path/filepath"
	"sync"

	"gitlab.markany.com/argos/cleaner/excludes"
	"gitlab.markany.com/argos/cleaner/fileinfo"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Watch : root 와 하위디렉토리 내의 파일/디렉토리의 생성 삭제를 감시
// 파일 생성 notification 만 ch 로 보냄
// 하위 디렉토리는 감지 목록에 자동으로 추가/제거됨
func Watch(root string, ch chan fileinfo.FileInfo) error {
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
		return err
	}

	// 이벤트 리시버
	go func() {
		defer watcher.Close()
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Debugf("EVENT: %s, %s\n", event.Name, event.Op)

				// Check the name if it is excluded map
				if excludes.Contains(event.Name) {
					continue
				}

				// Create
				if event.Op&fsnotify.Create == fsnotify.Create {
					if fi, err := os.Lstat(event.Name); err != nil {
						log.Error(err)
					} else {
						if fi.IsDir() {
							if err := watcher.Add(event.Name); err == nil {
								// 감지목록에 추가되기 전에 만들어진 것들 체크
								// 하위 디렉토리나 파일이 있을 경우 추가해줘야 함
								go watchDirOrAddFile(watcher, event.Name, ch)
							} else {
								log.Warn(err)
							}
						} else {
							ch <- fileinfo.FileInfo{
								Path: event.Name,
								Time: fi.ModTime(),
							}
						}
					}
				}

				// Remove from watcher
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if err := watcher.Remove(event.Name); err != nil {
						log.Debug(err)
					}
				}

			// watch for errors
			case err := <-watcher.Errors:
				log.Errorln("ERROR", err)
			}
		}
	}()

	// 감시할 디렉토리 추가
	if err := filepath.Walk(root, watchDir(watcher)); err != nil {
		log.Errorln("ERROR", err)
		return err
	}
	return nil
}

func watchDirOrAddFile(watcher *fsnotify.Watcher, path string, ch chan fileinfo.FileInfo) {
	deleteHidden := viper.GetBool("delete-hidden")
	base := filepath.Base(path)

	err := filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
		}
		// 하위폴더를 와처에 추가
		if fi.Mode().IsDir() {
			if base[0] == '.' && deleteHidden == false {
				return filepath.SkipDir
			}
			err := watcher.Add(path)
			if err != nil {
				return filepath.SkipDir
			}
		} else {
			ch <- fileinfo.FileInfo{
				Path: path,
				Time: fi.ModTime(),
			}
		}
		return nil
	})

	if err != nil {
		log.Error(err)
	}
	return
}

// watchDir, WalkFunc, 와처에 디렉토리 추가
func watchDir(watcher *fsnotify.Watcher) filepath.WalkFunc {
	deleteHidden := viper.GetBool("delete-hidden")
	return func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
		}
		base := filepath.Base(path)
		// 하위폴더를 와처에 추가
		if fi.Mode().IsDir() {
			if base[0] == '.' && deleteHidden == false {
				return filepath.SkipDir
			}
			err := watcher.Add(path)
			if err != nil {
				return filepath.SkipDir
			}
		}
		return nil
	}
}

// HandleCreate
func HandleFileCreation(chWatcher <-chan fileinfo.FileInfo, qFilesWatched *list.List, mutexQ *sync.Mutex) {
	for {
		select {
		// watch for fileInfo
		case fi := <-chWatcher:
			//log.Infof("%s", fi)
			// TODO: 성능문제 생기면 priorityQueue 에 저장
			if len(chWatcher) > 0 && len(chWatcher)%5 == 0 {
				log.Warn("event len = ", len(chWatcher))
			}
			mutexQ.Lock()
			qFilesWatched.PushBack(&fi)
			mutexQ.Unlock()
		}
	}
}
