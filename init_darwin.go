package main

import (
	"fmt"
	"syscall"
)

func init() {
	// increase file limit
	// https://stackoverflow.com/questions/17817204/how-to-set-ulimit-n-from-a-golang-program
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting rlimit ", err)
		return //os.Exit(1)
	}
	//fmt.Println("Current rLimit:", rLimit)

	//rLimit.Max = 20480
	rLimit.Cur = 20480
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
	}
}
