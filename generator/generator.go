package main

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const appName = "generator"

// Init function, runs before main()
func init() {
	// Read command line flags
	users := flag.Int("users", 1, "Concurrent user count, 동접 유저")
	duration := flag.Int("duration", 30, "Test duration in seconds")
	interval := flag.String("interval", "1s", `Interval for file creation("ns", "us" (or "µs"), "ms", "s", "m", "h")`)
	size := flag.Int64("size", 1024, "Size of file to create in bytes")
	path := flag.String("path", "", "[Required] Path where to create files")

	flag.Parse()

	parsedInterval, err := time.ParseDuration(*interval)
	if err != nil || *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Pass the flag values into viper.
	viper.Set("users", *users)
	viper.Set("duration", *duration)
	viper.Set("interval", parsedInterval)
	viper.Set("size", *size)
	viper.Set("path", *path)

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Infof("Starting %v...\n", appName)

	os.Mkdir(viper.GetString("path"), os.ModePerm)

	wg := sync.WaitGroup{}

	for i := 0; i < viper.GetInt("users"); i++ {
		wg.Add(1)
		durationChan := time.After(time.Duration(viper.GetInt("duration")) * time.Second)
		go Worker(i, &wg, durationChan)
	}

	wg.Wait()
}

func Worker(node int, wg *sync.WaitGroup, done <-chan time.Time) {
	defer wg.Done()
	log.Debugf("Worker[%d] started", node)

	dataDir, err := ioutil.TempDir(viper.GetString("path"),
		"data"+strconv.Itoa(node)+"-*")
	if err != nil {
		log.Error(err)
		return
	}

	intervalChan := time.Tick(viper.GetDuration("interval"))
	i := 0

	for {
		select {
		case <-intervalChan:
			f, err := ioutil.TempFile(dataDir, "file"+strconv.Itoa(i)+"-*")
			if err != nil {
				log.Error(err)
				return
			}
			i++

			f.Seek(int64(viper.GetInt64("size")-1), io.SeekCurrent)
			f.Write([]byte{0x1})
			f.Close()

			log.Debugf("[%d] %s", node, f.Name())
		case <-done:
			log.Debugf("Worker[%d] done", node)
			return
		}
	}
}

// Handles Ctrl+C or most other means of "controlled" shutdown gracefully. Invokes the supplied func before exiting.
func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}
