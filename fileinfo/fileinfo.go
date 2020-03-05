package fileinfo

import (
	"fmt"
	"time"
)

// FileInfo : 파일패스 와 생성한 시간 정보
type FileInfo struct {
	Path string
	Time time.Time
}

func (fi FileInfo) String() string {
	return fmt.Sprintf("%s: %s", fi.Path, fi.Time.String())
}
