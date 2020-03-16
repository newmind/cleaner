package diskinfo

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/disk"
)

const root = "./test"

func TestDiskUsage(t *testing.T) {
	disk := DiskUsage(root)
	t.Log(disk)
	fmt.Printf("All: %.2f GB\n", float64(disk.All)/float64(GB))
	fmt.Printf("Avail: %.2f GB\n", float64(disk.Avail)/float64(GB))
	fmt.Printf("Used: %.2f GB\n", float64(disk.Used)/float64(GB))
}

func TestDiskUsage_psutil(t *testing.T) {
	usage, _ := disk.Usage(root)
	t.Log(usage)
}
