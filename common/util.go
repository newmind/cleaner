package common

import (
	"os"
	"strconv"
)

func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func Atoi(s string, def int) int {
	if s == "" {
		return def
	}
	if ret, err := strconv.Atoi(s); err == nil {
		return ret
	}
	return def
}
