package main

import (
	"fmt"
	"log"
	"strings"
)

// State Machine Stuff
type State int

const (
	Init State = iota
	Start_Summ
	End_Summ
	Main_Log
	End_Log
	Exit
)

func initStates() map[State]func(string, Configuration, BuildInfo) State {
	return map[State]func(string, Configuration, BuildInfo) State{
		Init:       Init_State,
		Start_Summ: Start_Summ_State,
		End_Summ:   End_Summ_State,
		Main_Log:   Main_Log_State,
		End_Log:    End_Log_State,
	}
}

func Init_State(line string, conf Configuration, buildinfo BuildInfo) State {
	if strings.Contains(line, "-- START BUILD INFO --") {
		return Start_Summ
	} else {
		return Init
	}
}

func Start_Summ_State(line string, conf Configuration, buildinfo BuildInfo) State {

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
		return End_Summ
	}

	return Start_Summ
}

func formatBuildInfo(buildinfo BuildInfo, conf Configuration) {
	var info = buildinfo.Matches
	var builtLine string = fmt.Sprintf("START:   Requestor: %v, Project: %v, Def: %v", info["requestor"], info["projects"], info["builddef"])
	WriteToIrcBot(builtLine, conf)
}

func End_Summ_State(line string, conf Configuration, buildinfo BuildInfo) State {
	if strings.Contains(line, "-- END BUILD INFO --") {
		return Main_Log
	} else {
		return End_Summ
	}
}

func Main_Log_State(line string, conf Configuration, buildinfo BuildInfo) State {
	var builtLine string
	var info = buildinfo.Matches
	// Use HasPrefix instead of Contains in case of other exec'd ant jobs (like RTC tagging)
	if strings.HasPrefix(line, "BUILD SUCCESSFUL") {
		builtLine = fmt.Sprintf("SUCCESS: Requestor: %v, Project: %v, Def: %v", info["requestor"], info["projects"], info["builddef"])
		WriteToIrcBot(builtLine, conf)
		builtLine = formatBuildLogUrl(conf, buildinfo)
		WriteToIrcBot(builtLine, conf)
		return End_Log
	}
	if strings.HasPrefix(line, "BUILD FAILED") {
		builtLine = fmt.Sprintf("FAILED:  Requestor: %v, Project: %v, Def: %v", info["requestor"], info["projects"], info["builddef"])
		WriteToIrcBot(builtLine, conf)
		builtLine = formatBuildLogUrl(conf, buildinfo)
		WriteToIrcBot(builtLine, conf)
		return End_Log
	}
	return Main_Log
}

func formatBuildLogUrl(conf Configuration, build BuildInfo) string {
	var builtLine string = fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, build.Matches["uuid"])
	return builtLine
}

func End_Log_State(line string, conf Configuration, buildinfo BuildInfo) State {
	log.Println("Logfile finished")
	return Exit
}
