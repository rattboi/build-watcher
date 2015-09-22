package main

import (
	"./pubsub"
	"log"
)

func main() {
	log.Println("Starting build-watcher server")

	conf := handleConfig()

	done := make(chan bool)

	redisClient := pubsub.NewRedisClient("localhost")
	// Subscribe to redis event queue
	redisClient.Subscribe("queue")

	for {
		message := redisClient.Receive()

	}

	<-done
}

func newTailWatcher(states StateMapper, conf Configuration) {

	build := initBuildInfo()
	logState := initLog

	for {
		nextLogState := states[logState](string(b), build)

		if nextLogState != logState {
			finished := handleState(nextLogState, build, fileName, conf)
			if finished {
				return
			}
			logState = nextLogState
		}
	}
}

// returns true if at a "finished" state
func handleState(state State, build BuildInfo, fileName string, conf Configuration) bool {
	switch state {
	case mainLog:
		WriteToIrcBot(getBuildInfo("START", build), conf)
		return false
	case successLog:
		WriteToIrcBot(getBuildInfo("SUCCESS", build), conf)
		log.Printf("Logfile finished - SUCCESS - %v", fileName)
		return true
	case failLog:
		var fail = getBuildInfo("FAIL", build) + formatBuildLogUrl(build, conf)
		WriteToIrcBot(fail, conf)
		log.Printf("Logfile finished - FAIL - %v", fileName)
		return true
	case abandonLog:
		WriteToIrcBot(getBuildInfo("ABANDON", build), conf)
		log.Printf("Logfile finished - ABANDON - %v", fileName)
		return true
	case exitLog:
		log.Printf("Logfile finished - EXIT - %v", fileName)
		return true
	default:
		return false
	}
}
