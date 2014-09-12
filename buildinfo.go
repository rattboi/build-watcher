package main

import (
	"fmt"
	"regexp"
	"strings"
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

func getBuildInfo(buildStatus string, buildInfo BuildInfo) string {
	var info = buildInfo.Matches
	switch buildStatus {
	case "START":
		{
			var res []string
			var envRE, _ = regexp.Compile(`Deploy to (.*) - DEPLOY ONE PROJECT.*`)
			res = envRE.FindStringSubmatch(info["builddef"])
			var parsedProjects = strings.Replace(info["projects"], "EIT_1World_", "", -1)
			parsedProjects = strings.Replace(parsedProjects, "EIT_1WORLD_", "", -1)
			if res != nil {
				return fmt.Sprintf("%v: %v : %v : %v : %v ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"],14), pad(res[1], 12), parsedProjects)
			} else {
				return fmt.Sprintf("%v: %v : %v : %v : %v ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"],14), info["builddef"], parsedProjects)
			}
		}
    case "FAIL":
        {
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), colorMsg(pad(buildStatus, 7), cBlack, cRed), pad(info["requestor"], 14))
        }
    case "SUCCESS":
        {
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), colorMsg(pad(buildStatus, 7), cWhite, cGreen), pad(info["requestor"], 14))
        }
	default:
		{
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"], 12))
		}
	}
}

func formatBuildLogUrl(buildInfo BuildInfo, conf Configuration) string {
	var info = buildInfo.Matches
	return fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, info["uuid"])
}
