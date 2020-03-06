package excludes

import "path/filepath"

var (
	// 스캔, 파일 감지에서 제외할 파일/디렉토리 목록
	excludes = map[string]bool{}
)

func init() {
	excludeFromWatch := []string{
		".fseventsd", // MacOS ramdisk system directory
		".Trashes",   //
	}

	for _, e := range excludeFromWatch {
		Add(e)
	}
}

func Add(path string) {
	path = filepath.Clean(path)
	excludes[path] = true
}

func Remove(path string) {
	path = filepath.Clean(path)
	delete(excludes, path)
}

func Contains(path string) bool {
	path = filepath.Clean(path)
	return excludes[path]
}
