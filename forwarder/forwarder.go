package main

import (
	"../pubsub"
	"bufio"
	"fmt"
	"gopkg.in/fsnotify.v0"
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

func main() {
	log.Println("Starting build-watcher")

	conf := handleConfig()

	log.Println("Creating fsnotify watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	// Process events
	log.Println("Starting to process events")
	go processEvents(watcher, conf)

	log.Println("Adding watchdir: ", conf.Watchdir)
	err = watcher.Watch(conf.Watchdir)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func processEvents(watcher *fsnotify.Watcher, conf Configuration) {
	for {
		select {
		case ev := <-watcher.Event:
			// See if a new file is created that matches our pattern
			if ev.IsCreate() && isLogFile(ev.Name, conf.Filepattern) {
				// 'fork' a new tail watcher for this build file
				go newTailWatcher(ev.Name, conf)
			}
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}

func isLogFile(fileName string, filePattern string) bool {
	log.Println("New File -> ", fileName)
	re, _ := regexp.Compile(filePattern)
	res := re.FindStringSubmatch(fileName)
	if res == nil {
		log.Println("No match to " + filePattern)
		return false
	}
	return true
}

func newTailWatcher(fileName string, conf Configuration) {

	redisClient := pubsub.NewRedisClient("localhost")

	var in io.Reader
	fin, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't open file: "+err.Error())
	}
	defer fin.Close()
	fin.Seek(0, os.SEEK_END)
	in = fin
	buf := bufio.NewReader(in)

	for {
		b, _, err := buf.ReadLine()
		if len(b) > 0 {
			// send to redis "queue" here
			redisClient.Publish("queue", string(b))
		}
		if err != nil {
			if err != io.EOF {
				return
			}
			// if the file we're watching no longer exists, we should stop watching it
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
