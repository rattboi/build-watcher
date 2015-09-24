package main

import (
	"../pubsub"
	"log"
)

func main() {
	log.Println("Starting build-watcher server")

	conf := handleConfig()

	hostClient := pubsub.NewRedisClient("localhost")
	// Subscribe to redis event queue
	hostClient.Subscribe("hosts")

	for {
		host := hostClient.Receive()
		go processHostSubscription(host, conf)
	}
}

func processHostSubscription(host pubsub.Message, conf Configuration) {

	watcherClient := pubsub.NewRedisClient("localhost")
	watcherClient.Subscribe(host.Data)

	states := initStates()
	build := initBuildInfo()
	logState := initLog

	for {
		logMessage := watcherClient.Receive()
		logLine := logMessage.Data

		nextLogState := states[logState](logLine, build)

		if nextLogState != logState {
			finished := handleState(nextLogState, build, conf)
			if finished {
				return
			}
			logState = nextLogState
		}
	}
}

// returns true if at a "finished" state
func handleState(state State, build BuildInfo, conf Configuration) bool {
	switch state {
	case mainLog:
		WriteToIrcBot(getBuildInfo("START", build), conf)
		return false
	case successLog:
		WriteToIrcBot(getBuildInfo("SUCCESS", build), conf)
		return true
	case failLog:
		var fail = getBuildInfo("FAIL", build) + formatBuildLogUrl(build, conf)
		WriteToIrcBot(fail, conf)
		return true
	case abandonLog:
		WriteToIrcBot(getBuildInfo("ABANDON", build), conf)
		return true
	case exitLog:
		return true
	default:
		return false
	}
}
