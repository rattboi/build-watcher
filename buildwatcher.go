package main

import (
	"log"
	"net"
	"regexp"

	"github.com/ActiveState/tail"
	"github.com/howeyc/fsnotify"
)

// IRC Bot Helper
func WriteToIrcBot(message string, conf Configuration) {
	strEcho := message + "\n"
	servAddr := conf.Botaddress + ":" + conf.Botport
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)

	if err != nil {
		log.Println("ResolveTCPAddr failed:", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()

	if err != nil {
		log.Println("Dial failed:", err.Error())
		return
	}

	_, err = conn.Write([]byte(strEcho))

	if err != nil {
		log.Println("Write to server failed:", err.Error())
		return
	}
}

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
				if isLogFile(ev.Name, conf) && ev.IsCreate() {
					// 'fork' a new tail watcher for this build file
					go func() {
						t, err := tail.TailFile(ev.Name, tail.Config{Follow: true})
						if err != nil {
							return
						}

						var build BuildInfo
						initBuildInfo(&build)

						logState := Init
						for line := range t.Lines {
							logState = states[logState](line.Text, conf, build)
							if logState == Exit {
								return
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
