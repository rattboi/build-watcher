package main

import (
	"fmt"
	"regexp"
)

// Build Info Stuff
type BuildInfo struct {
	Patterns map[string]*regexp.Regexp
	Matches  map[string]string
}

func initBuildInfo(build *BuildInfo) {
	build.Patterns = make(map[string]*regexp.Regexp)
	build.Matches = make(map[string]string)

	build.Patterns["uuid"], _ = regexp.Compile(`.*Build Result UUID: (.*)`)
	build.Patterns["requestor"], _ = regexp.Compile(`.*Requestor ID: (.*)`)
	build.Patterns["enghost"], _ = regexp.Compile(`.*Build Engine Host: (.*)`)
	build.Patterns["engid"], _ = regexp.Compile(`.*Build Engine ID: (.*)`)
	build.Patterns["builddef"], _ = regexp.Compile(`.*Build Definition: (.*)`)
	build.Patterns["projects"], _ = regexp.Compile(`.*Project Names: (.*)`)
	build.Patterns["buildlabel"], _ = regexp.Compile(`.*Build Label: (.*)`)

	for k, _ := range build.Patterns {
		build.Matches[k] = ""
	}
}

func formatBuildInfo(buildStatus string, buildInfo BuildInfo) string {
	var info = buildInfo.Matches
	var builtLine string = fmt.Sprintf("%v: %-10v: Requestor: %v, Project: %v, Def: %v", info["buildlabel"], buildStatus, info["requestor"], info["projects"], info["builddef"])
	return builtLine
}

func formatBuildLogUrl(buildInfo BuildInfo, conf Configuration) string {
	var info = buildInfo.Matches
	var builtLine string = fmt.Sprintf("%v: %v/resource/itemOid/com.ibm.team.build.BuildResult/%v", info["buildlabel"], conf.RTCBaseURL, info["uuid"])
	return builtLine
}
