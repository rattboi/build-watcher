package main

import (
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
	go processEvents(watcher, initStates(), conf)

	log.Println("Adding watchdir: ", conf.Watchdir)
	err = watcher.Watch(conf.Watchdir)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func processEvents(watcher *fsnotify.Watcher, states StateMapper, conf Configuration) {
	for {
		select {
		case ev := <-watcher.Event:
			// See if a new file is created that matches our pattern
			if ev.IsCreate() && isLogFile(ev.Name, conf.Filepattern) {
				// 'fork' a new tail watcher for this build file
				go newTailWatcher(ev.Name, states, conf)
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

func newTailWatcher(fileName string, states StateMapper, conf Configuration) {

	build := initBuildInfo()
	logState := initLog

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
			nextLogState := states[logState](string(b), build)

			if nextLogState != logState {
				finished := handleState(nextLogState, build, fileName, conf)
				if finished {
					return
				}
				logState = nextLogState
			}
		}
		if err != nil {
			if err != io.EOF {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// returns true if at a "finished" state
func handleState(state State, build BuildInfo, fileName string, conf Configuration) bool {
	switch state {
	case mainLog:
		WriteToBot(getBuildInfo("START", build, conf), conf)
		return false
	case successLog:
		WriteToBot(getBuildInfo("SUCCESS", build, conf), conf)
		log.Printf("Logfile finished - SUCCESS - %v", fileName)
		return true
	case failLog:
		var fail = getBuildInfo("FAIL", build, conf) // + formatBuildLogUrl(build, conf)
		WriteToBot(fail, conf)
		log.Printf("Logfile finished - FAIL - %v", fileName)
		return true
	case abandonLog:
		WriteToBot(getBuildInfo("ABANDON", build, conf), conf)
		log.Printf("Logfile finished - ABANDON - %v", fileName)
		return true
	case exitLog:
		log.Printf("Logfile finished - EXIT - %v", fileName)
		return true
	default:
		return false
	}
}
