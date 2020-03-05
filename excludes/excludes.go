package excludes

var (
	// 스캔, 파일 감지에서 제외할 파일/디렉토리 목록
	Excludes         = map[string]bool{}
	ExcludeFromWatch = []string{
		".fseventsd", // MacOS ramdisk system directory
		".Trashes",   //
	}
)

func init() {
	for _, e := range ExcludeFromWatch {
		Excludes[e] = true
	}
}
