package scanner

func FpW(n int, root string) {
	for i := 0; i < n; i++ {
		FilePathWalkDir(root)
	}
}
func IoRdir(n int, root string) {
	for i := 0; i < n; i++ {
		IOReadDir(root)
	}
}
func OsRdir(n int, root string) {
	for i := 0; i < n; i++ {
		OSReadDir(root)
	}
}
