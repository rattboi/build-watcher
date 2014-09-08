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

	for k, _ := range build.Patterns {
		build.Matches[k] = ""
	}
}

func formatBuildInfo(buildStatus string, buildinfo BuildInfo) string {
	var info = buildinfo.Matches
	var builtLine string = fmt.Sprintf("%-10v: Requestor: %v, Project: %v, Def: %v", buildStatus, info["requestor"], info["projects"], info["builddef"])
	return builtLine
}

func formatBuildLogUrl(build BuildInfo, conf Configuration) string {
	var builtLine string = fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, build.Matches["uuid"])
	return builtLine
}
