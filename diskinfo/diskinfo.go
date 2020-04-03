package diskinfo

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
)

type DiskStatus struct {
	All   uint64 `json:"all"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Avail uint64 `json:"avail"`
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func unescapeFstab(path string) string {
	escaped, err := strconv.Unquote(`"` + path + `"`)
	if err != nil {
		return path
	}
	return escaped
}

func intToString(orig []int8) string {
	ret := make([]byte, len(orig))
	size := -1
	for i, o := range orig {
		if o == 0 {
			size = i
			break
		}
		ret[i] = byte(o)
	}
	if size == -1 {
		size = len(orig)
	}

	return string(ret[0:size])
}

func uintToString(orig []uint8) string {
	ret := make([]byte, len(orig))
	size := -1
	for i, o := range orig {
		if o == 0 {
			size = i
			break
		}
		ret[i] = byte(o)
	}
	if size == -1 {
		size = len(orig)
	}

	return string(ret[0:size])
}

func GetAllPartitions() ([]disk.PartitionStat, error) {
	pp, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	sort.Slice(pp, func(i, j int) bool {
		return len(pp[i].Mountpoint) > len(pp[j].Mountpoint)
	})
	return pp, err
}

func GetMountpoint(dir string) string {
	pp, err := GetAllPartitions()
	if err != nil {
		log.Error(err)
		return ""
	}
	dir = filepath.Clean(dir)
	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Error(err)
		return ""
	}
	for _, p := range pp {
		if strings.HasPrefix(dir, p.Mountpoint) {
			return p.Mountpoint
		}
	}
	return ""
}
