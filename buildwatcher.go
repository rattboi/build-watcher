package main

import (
	"log"
	"regexp"

	"github.com/ActiveState/tail"
	"github.com/howeyc/fsnotify"
)

func isLogFile(filename string, conf Configuration) bool {
	log.Println("New File -> ", filename)
	re, _ := regexp.Compile(conf.Filepattern)
	res := re.FindStringSubmatch(filename)
	if res == nil {
		log.Println("No match to " + conf.Filepattern)
		return false
	}
	return true
}

func main() {
	log.Println("Starting build-watcher")

	var conf Configuration
	// set some dumb defaults
	setConfigDefaults(&conf)
	// get settings from config file if it exists and override defaults
	parseConfig(&conf)
	// parse out any flags and override defaults/config
	setupFlags(&conf)

	states := initStates()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				// See if a new file is created that matches our pattern
				if ev.IsCreate() && isLogFile(ev.Name, conf) {
					// 'fork' a new tail watcher for this build file
					go func() {
						t, err := tail.TailFile(ev.Name, tail.Config{Follow: true})
						if err != nil {
							return
						}

						var build BuildInfo
						initBuildInfo(&build)

						logState := initLog

						for line := range t.Lines {
							nextLogState := states[logState](line.Text, build)

							if nextLogState != logState {
								// log.Printf("State Transition: %v -> %v\n", logState, nextLogState)
								switch nextLogState {
								case mainLog:
									WriteToIrcBot(getBuildInfo("START", build), conf)
								case successLog:
									WriteToIrcBot(getBuildInfo("SUCCESS", build), conf)
								case failLog:
									var fail = getBuildInfo("FAIL", build) + formatBuildLogUrl(build, conf)
									WriteToIrcBot(fail, conf)
								case abandonLog:
									WriteToIrcBot(getBuildInfo("ABANDON", build), conf)
								case exitLog:
									log.Println("Logfile finished")
									return
								default:
								}

								logState = nextLogState
							}
						}
					}()
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(conf.Watchdir)
	if err != nil {
		log.Fatal(err)
	}

	<-done

	watcher.Close()
}
