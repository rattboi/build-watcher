package main

import (
	"log"
	"strings"
)

// State Machine Stuff
type State int

const (
	initLog State = iota
	startLog
	startSumm
	endSumm
	mainLog
	successLog
	failLog
	endLog
	exitLog
)

func initStates() map[State]func(string, BuildInfo) State {
	return map[State]func(string, BuildInfo) State{
		initLog:    initLogState,
		startLog:   startLogState,
		startSumm:  startSummState,
		endSumm:    endSummState,
		mainLog:    mainLogState,
		successLog: successLogState,
		failLog:    failLogState,
		endLog:     endLogState,
	}
}

func initLogState(line string, buildinfo BuildInfo) State {
	return startLog
}

func startLogState(line string, buildinfo BuildInfo) State {
	if strings.Contains(line, "-- START BUILD INFO --") {
		return startSumm
	} else if strings.HasPrefix(line, "BUILD ") { // This is a deploy not using StreamBuild, or failed very early. Just allow closing the tail log
		line = strings.TrimPrefix(line, "BUILD ")
		if strings.Contains(line, "SUCCESS") {
			return endLog
		}
		if strings.Contains(line, "FAIL") {
			return endLog
		}
	}
	return startLog
}

func startSummState(line string, buildinfo BuildInfo) State {
	var res []string

	for k, v := range buildinfo.Patterns {
		res = v.FindStringSubmatch(line)
		if res != nil {
			buildinfo.Matches[k] = res[1]
		}
	}

	var allMatched bool = true
	for _, v := range buildinfo.Matches {
		if v == "" {
			allMatched = false
		}
	}

	if allMatched {
		return endSumm
	} else {
		return startSumm
	}
}

func endSummState(line string, buildinfo BuildInfo) State {
	if strings.Contains(line, "-- END BUILD INFO --") {
		return mainLog
	} else {
		return endSumm
	}
}

func mainLogState(line string, buildinfo BuildInfo) State {
	// Use HasPrefix instead of Contains in case of other exec'd ant jobs (like RTC tagging)
	if strings.HasPrefix(line, "BUILD ") {
		line = strings.TrimPrefix(line, "BUILD ")
		if strings.Contains(line, "SUCCESS") {
			return successLog
		}
		if strings.Contains(line, "FAIL") {
			return failLog
		}
	}
	return mainLog
}

func successLogState(line string, buildinfo BuildInfo) State {
	return endLog
}

func failLogState(line string, buildinfo BuildInfo) State {
	return endLog
}

func endLogState(line string, buildinfo BuildInfo) State {
	log.Println("Logfile finished")
	return exitLog
}
