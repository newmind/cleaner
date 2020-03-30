package diskinfo

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/shirou/gopsutil/disk"
)

const root = "./test"

//func TestDiskUsage(t *testing.T) {
//	disk := DiskUsage(root)
//	t.Log(disk)
//	fmt.Printf("All: %.2f GB\n", float64(disk.All)/float64(GB))
//	fmt.Printf("Avail: %.2f GB\n", float64(disk.Avail)/float64(GB))
//	fmt.Printf("Used: %.2f GB\n", float64(disk.Used)/float64(GB))
//}

func TestDiskUsage_psutil(t *testing.T) {
	usage, _ := disk.Usage(root)
	t.Log(usage)
}

func TestPartitions(t *testing.T) {
	pp, _ := disk.Partitions(true)
	sort.Slice(pp, func(i, j int) bool {
		return len(pp[i].Mountpoint) > len(pp[j].Mountpoint)
	})
	for _, p := range pp {
		t.Logf(p.String())
	}
	s, err := filepath.Abs("")
	t.Log(s, err)
}
