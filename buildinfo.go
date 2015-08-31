package main

import (
	"fmt"
	"regexp"
	"strings"
)

var ColorToUse Color = cWhite

// Build Info Stuff
type BuildInfo struct {
	Patterns   map[string]*regexp.Regexp
	Matches    map[string]string
	labelColor Color
}

func initBuildInfo() BuildInfo {
	var build BuildInfo
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

	build.labelColor = ColorToUse
	ColorToUse = (ColorToUse + 1) % 16

	return build
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

func buildInfoString(info BuildInfo, status string, env string, proj string, fgColor Color, bgColor Color) string {
	return fmt.Sprintf("%v: %v : %v : %v : %v ", colorMatchedMsg(info.Matches["buildlabel"], info.labelColor), colorMsg(pad(status, 7), fgColor, bgColor), pad(info.Matches["requestor"], 14), pad(env, 14), summarizeProject(proj))
}

func getBuildInfo(buildStatus string, buildInfo BuildInfo) string {
	var info = buildInfo.Matches
	switch buildStatus {
	case "START":
		{
			var res1, res2 []string
			var batchEnvRE = regexp.MustCompile(`Deploy to (.*) - DEPLOY ONE PROJECT.*`)
			var selfSEnvRE = regexp.MustCompile(`(?i)Dev.* DEPLOY (.*?) - (.*)`)
			res1 = batchEnvRE.FindStringSubmatch(info["builddef"])
			res2 = selfSEnvRE.FindStringSubmatch(info["builddef"])
			if res1 != nil {
				return buildInfoString(buildInfo, buildStatus, res1[1], info["projects"], cWhite, cBlack)
			} else if res2 != nil {
				return buildInfoString(buildInfo, buildStatus, "SS - "+res2[1], res2[2], cWhite, cBlack)
			} else {
				return buildInfoString(buildInfo, buildStatus, info["builddef"], info["projects"], cWhite, cBlack)
			}
		}
	case "FAIL":
		{
			return buildInfoString(buildInfo, buildStatus, "", "", cBlack, cRed)
		}
	case "SUCCESS":
		{
			return buildInfoString(buildInfo, buildStatus, "", "", cBlack, cGreen)
		}
	case "ABANDON":
		{
			return buildInfoString(buildInfo, buildStatus, "", "", cWhite, cBrown)
		}
	default:
		{
			return buildInfoString(buildInfo, buildStatus, "", "", cWhite, cBlack)
		}
	}
}

func formatBuildLogUrl(buildInfo BuildInfo, conf Configuration) string {
	var info = buildInfo.Matches
	return fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, info["uuid"])
}
