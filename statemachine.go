package main

import (
	"fmt"
	"log"
	"strings"
)

// State Machine Stuff
type State int

const (
	initLog State = iota
	startSumm
	endSumm
	mainLog
	endLog
	exitLog
)

func initStates() map[State]func(string, Configuration, BuildInfo) State {
	return map[State]func(string, Configuration, BuildInfo) State{
		initLog:   initLogState,
		startSumm: startSummState,
		endSumm:   endSummState,
		mainLog:   mainLogState,
		endLog:    endLogState,
	}
}

func initLogState(line string, conf Configuration, buildinfo BuildInfo) State {
	if strings.Contains(line, "-- START BUILD INFO --") {
		return startSumm
	} else {
		return initLog
	}
}

func startSummState(line string, conf Configuration, buildinfo BuildInfo) State {

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
		formatBuildInfo(buildinfo, conf)
		return endSumm
	}

	return startSumm
}

func formatBuildInfo(buildinfo BuildInfo, conf Configuration) {
	var info = buildinfo.Matches
	var builtLine string = fmt.Sprintf("START:   Requestor: %v, Project: %v, Def: %v", info["requestor"], info["projects"], info["builddef"])
	WriteToIrcBot(builtLine, conf)
}

func endSummState(line string, conf Configuration, buildinfo BuildInfo) State {
	if strings.Contains(line, "-- END BUILD INFO --") {
		return mainLog
	} else {
		return endSumm
	}
}

func mainLogState(line string, conf Configuration, buildinfo BuildInfo) State {
	var builtLine string
	var info = buildinfo.Matches
	// Use HasPrefix instead of Contains in case of other exec'd ant jobs (like RTC tagging)
	if strings.HasPrefix(line, "BUILD ") {
		line = strings.TrimPrefix(line, "BUILD ")
		builtLine = fmt.Sprintf("%10v: Requestor: %v, Project: %v, Def: %v", line, info["requestor"], info["projects"], info["builddef"])
		WriteToIrcBot(builtLine, conf)
		builtLine = formatBuildLogUrl(conf, buildinfo)
		WriteToIrcBot(builtLine, conf)
		return endLog
	}
	return mainLog
}

func formatBuildLogUrl(conf Configuration, build BuildInfo) string {
	var builtLine string = fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, build.Matches["uuid"])
	return builtLine
}

func endLogState(line string, conf Configuration, buildinfo BuildInfo) State {
	log.Println("Logfile finished")
	return exitLog
}
