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

func summarizeProject(projects string) string {
	var parsedProjects = strings.Replace(projects, "EIT_1World_", "", -1)
	parsedProjects = strings.Replace(parsedProjects, "EIT_1WORLD_", "", -1)
	parsedProjects = strings.Replace(parsedProjects, "Service", "Svc", -1)
	parsedProjects = strings.Replace(parsedProjects, "Invoice", "Inv", -1)
	parsedProjects = strings.Replace(parsedProjects, "Entity", "Ent", -1)
	parsedProjects = strings.Replace(parsedProjects, "Custom", "Cust", -1)
	parsedProjects = strings.Replace(parsedProjects, "Operation", "Op", -1)
	parsedProjects = strings.Replace(parsedProjects, "Process", "Proc", -1)
	parsedProjects = strings.Replace(parsedProjects, "Preference", "Pref", -1)
	parsedProjects = strings.Replace(parsedProjects, "Schedule", "Sched", -1)
	return parsedProjects
}

func getBuildInfo(buildStatus string, buildInfo BuildInfo) string {
	var info = buildInfo.Matches
	switch buildStatus {
	case "START":
		{
			var res1, res2 []string
			var batchEnvRE = regexp.MustCompile(`Deploy to (.*) - DEPLOY ONE PROJECT.*`)
			var selfSEnvRE = regexp.MustCompile(`Dev DEPLOY (.*) - (.*)`)
			res1 = batchEnvRE.FindStringSubmatch(info["builddef"])
			res2 = selfSEnvRE.FindStringSubmatch(info["builddef"])
			if res1 != nil {
				return fmt.Sprintf("%v: %v : %v : %v : %v ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"], 14), pad(res1[1], 14), summarizeProject(info["projects"]))
			} else if res2 != nil {
				return fmt.Sprintf("%v: %v : %v : %v : %v ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"], 14), pad(res2[1], 14), summarizeProject(res2[2]))
			} else {
				return fmt.Sprintf("%v: %v : %v : %v : %v ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"], 14), info["builddef"], summarizeProject(info["projects"]))
			}
		}
	case "FAIL":
		{
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), colorMsg(pad(buildStatus, 7), cBlack, cRed), pad(info["requestor"], 14))
		}
	case "SUCCESS":
		{
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), colorMsg(pad(buildStatus, 7), cBlack, cGreen), pad(info["requestor"], 14))
		}
	case "ABANDON":
		{
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), colorMsg(pad(buildStatus, 7), cWhite, cBrown), pad(info["requestor"], 14))
		}
	default:
		{
			return fmt.Sprintf("%v: %v : %v : ", hashedColor(info["buildlabel"]), pad(buildStatus, 7), pad(info["requestor"], 14))
		}
	}
}

func formatBuildLogUrl(buildInfo BuildInfo, conf Configuration) string {
	var info = buildInfo.Matches
	return fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, info["uuid"])
}
