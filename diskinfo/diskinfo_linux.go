package diskinfo

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	fs := unix.Statfs_t{}
	err := unix.Statfs(path, &fs)
	if err != nil {
		return
	}
	// log.Infof("%s %s %s\n", unescapeFstab(path), intToString(fs.Mntonname[:]), intToString(fs.Mntfromname[:]))
	logrus.Infof("%s %#v\n", path, fs)
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Avail = fs.Bavail * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}
