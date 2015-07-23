package main

import (
	"log"
	"regexp"

	"github.com/ActiveState/tail"
	"gopkg.in/fsnotify.v0"
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

func newTailWatcher(filename string, states StateMapper, conf Configuration) {
	t, err := tail.TailFile(filename, tail.Config{Follow: true})
	if err != nil {
		return
	}

	build := initBuildInfo()
	logState := initLog

	for line := range t.Lines {
		nextLogState := states[logState](line.Text, build)

		if nextLogState != logState {
			// log.Printf("%v : Trans: %v -> %v\n", ev.Name, logState, nextLogState)
			handleState(nextLogState, build, conf)
			logState = nextLogState
		}
	}
}

func handleState(state State, build BuildInfo, conf Configuration) {
	switch state {
	case mainLog:
		WriteToIrcBot(getBuildInfo("START", build), conf)
	case successLog:
		WriteToIrcBot(getBuildInfo("SUCCESS", build), conf)
		log.Println("Logfile finished")
		return
	case failLog:
		var fail = getBuildInfo("FAIL", build) + formatBuildLogUrl(build, conf)
		WriteToIrcBot(fail, conf)
		log.Println("Logfile finished")
		return
	case abandonLog:
		WriteToIrcBot(getBuildInfo("ABANDON", build), conf)
		log.Println("Logfile finished")
		return
	case exitLog:
		log.Println("Logfile finished")
		return
	default:
	}
}
