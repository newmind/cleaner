package main

import (
	"fmt"
	"runtime"
	"syscall"
)

func init() {

	// increase file limit
	if runtime.GOOS == "linux" {
		// https://stackoverflow.com/questions/17817204/how-to-set-ulimit-n-from-a-golang-program
		var rLimit syscall.Rlimit
		err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			fmt.Println("Error Getting rlimit ", err)
			return //os.Exit(1)
		}
		//fmt.Println("Current rLimit:", rLimit)

		rLimit.Max = 900000
		rLimit.Cur = 900000
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			fmt.Println("Error Setting Rlimit ", err)
		}
	}
}
