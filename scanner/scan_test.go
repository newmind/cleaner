package scanner

import (
	"fmt"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
)

var root string = "../test"

// var root string = "/Volumes/RAMDisk"

func Test_visit(t *testing.T) {
	var files []string

	// scanner(root, )
	root, _ := filepath.EvalSymlinks(root)

	err := filepath.Walk(root, visit(&files))
	if err != nil {
		panic(err)
	}
	fmt.Println(len(files))
	// for _, file := range files {
	// 	fmt.Println(file)
	// }
}

func Test_FilePathWalkDir(t *testing.T) {
	files, err := FilePathWalkDir(root)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(files))
	// for _, file := range files {
	// 	fmt.Println(file)
	// }
}

func Test_GoDirWalk(t *testing.T) {
	files, err := GoDirWalk(root)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(files))
	// for _, file := range files {
	// 	fmt.Println(file)
	// }
}

func Test_IOReadDir(t *testing.T) {
	files, err := IOReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
}

func Test_OSReadDir(t *testing.T) {
	files, err := OSReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
}

func Benchmark_GoDirWalk(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GoDirWalk(root)
	}
}

func Benchmark_FilePathWalkDir1(b *testing.B) { benchmarkFilePathWalkDir(1, root, b) }

// func Benchmark_FilePathWalkDir2(b *testing.B)  { benchmarkFilePathWalkDir(2, root, b) }
// func Benchmark_FilePathWalkDir3(b *testing.B)  { benchmarkFilePathWalkDir(3, root, b) }
// func Benchmark_FilePathWalkDir10(b *testing.B) { benchmarkFilePathWalkDir(10, root, b) }
// func Benchmark_FilePathWalkDir20(b *testing.B) { benchmarkFilePathWalkDir(20, root, b) }
// func Benchmark_FilePathWalkDir30(b *testing.B) { benchmarkFilePathWalkDir(30, root, b) }
// func Benchmark_FilePathWalkDir40(b *testing.B) { benchmarkFilePathWalkDir(40, root, b) }
// func Benchmark_FilePathWalkDir50(b *testing.B) { benchmarkFilePathWalkDir(50, root, b) }

func Benchmark_IOReadDir1(b *testing.B) { benchmarkIOReadDir(1, root, b) }

// func Benchmark_IOReadDir2(b *testing.B)        { benchmarkIOReadDir(2, root, b) }
// func Benchmark_IOReadDir3(b *testing.B)        { benchmarkIOReadDir(3, root, b) }
// func Benchmark_IOReadDir10(b *testing.B)       { benchmarkIOReadDir(10, root, b) }
// func Benchmark_IOReadDir20(b *testing.B)       { benchmarkIOReadDir(20, root, b) }
// func Benchmark_IOReadDir30(b *testing.B)       { benchmarkIOReadDir(30, root, b) }
// func Benchmark_IOReadDir40(b *testing.B)       { benchmarkIOReadDir(40, root, b) }
// func Benchmark_IOReadDir50(b *testing.B)       { benchmarkIOReadDir(50, root, b) }

func Benchmark_OSReadDir1(b *testing.B) { benchmarkOSReadDir(1, root, b) }

// func Benchmark_OSReadDir2(b *testing.B)        { benchmarkOSReadDir(2, root, b) }
// func Benchmark_OSReadDir3(b *testing.B)        { benchmarkOSReadDir(3, root, b) }
// func Benchmark_OSReadDir10(b *testing.B)       { benchmarkOSReadDir(10, root, b) }
// func Benchmark_OSReadDir20(b *testing.B)       { benchmarkOSReadDir(20, root, b) }
// func Benchmark_OSReadDir30(b *testing.B)       { benchmarkOSReadDir(30, root, b) }
// func Benchmark_OSReadDir40(b *testing.B)       { benchmarkOSReadDir(40, root, b) }
// func Benchmark_OSReadDir50(b *testing.B)       { benchmarkOSReadDir(50, root, b) }

func benchmarkFilePathWalkDir(i int, root string, b *testing.B) {
	// run the FilePathWalkDir function b.N times
	for n := 0; n < b.N; n++ {
		FpW(i, root)
	}
}
func benchmarkIOReadDir(i int, root string, b *testing.B) {
	// run the IOReadDir function b.N times
	for n := 0; n < b.N; n++ {
		IoRdir(i, root)
	}
}
func benchmarkOSReadDir(i int, root string, b *testing.B) {
	// run the OSReadDir function b.N times
	for n := 0; n < b.N; n++ {
		OsRdir(i, root)
	}
}
