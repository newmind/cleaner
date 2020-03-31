package main

import (
	"flag"
	"fmt"
	"gitlab.markany.com/argos/cleaner/vods"
	"os"
)

const appName = "listvod"

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Printf("Usage : %s path", appName)
		os.Exit(1)
	}

	list := vods.ListAllVODs(args[0])
	fmt.Printf("Total %d CCTVs", len(list))
}
