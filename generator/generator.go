package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func main() {
	f, err := os.OpenFile("./test/test", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.Seek(8, io.SeekCurrent)
	f.Write([]byte{0x1})
	f.Close()

	fi, err := os.Stat(f.Name())
	log.Info(fi.Size())

	log.Infof("%#v, name = %s", f, f.Name())
}
