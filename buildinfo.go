package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type SlackMsg struct {
	Channel     string        `json:"channel"`
	Username    string        `json:"username,omitempty"`
	Text        string        `json:"text,omitempty"`
	Attachments []Attachments `json:"attachments,omitempty"`
}

type Attachments struct {
	Fallback   string   `json:"fallback,omitempty"`
	Color      string   `json:"color,omitempty"`
	Pretext    string   `json:"pretext,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	AuthorIcon string   `json:"author_icon,omitempty"`
	Title      string   `json:"title,omitempty"`
	TitleLink  string   `json:"title_link,omitempty"`
	Text       string   `json:"text"`
	Fields     []Fields `json:"fields,omitempty"`
	ImageURL   string   `json:"image_url,omitempty"`
	ThumbURL   string   `json:"thumb_url,omitempty"`
}

type Fields struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func (m SlackMsg) Encode() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (m SlackMsg) Post(WebhookURL string) error {
	encoded, err := m.Encode()
	if err != nil {
		return err
	}

	resp, err := http.PostForm(WebhookURL, url.Values{"payload": {encoded}})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Not OK")
	}
	return nil
}

func WriteToBot(msg SlackMsg, conf Configuration) {
	if msg.Channel != "" {
		err := msg.Post(conf.WebhookUrl)
		if err != nil {
			log.Printf("Post failed: %v", err)
		}
	}
}

// Build Info Stuff
type BuildInfo struct {
	Patterns map[string]*regexp.Regexp
	Matches  map[string]string
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

func createColor(stringToHash string) string {
	h := fnv.New32a()
	h.Write([]byte(stringToHash))
	var x uint32 = h.Sum32()
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, x)
	s := hex.EncodeToString(buf)
	return s[:6]
}

func createBuildInfoMessage(info BuildInfo, status string, env string, proj string, conf Configuration) SlackMsg {
	formatted := fmt.Sprintf("%v\n%v : %v : %v : %v", info.Matches["buildlabel"], status, info.Matches["requestor"], env, summarizeProject(proj))
	newformatted := fmt.Sprintf(" %v - %v - %v", env, summarizeProject(proj), info.Matches["requestor"])

	colorHash := fmt.Sprintf("%v%v%v%v", info.Matches["buildlabel"], info.Matches["requestor"], env, summarizeProject(proj))
	color := createColor(colorHash)

	buildResultUrl := fmt.Sprintf("<%v/resource/itemOid/com.ibm.team.build.BuildResult/%v|Build Results>", conf.RTCBaseURL, info.Matches["uuid"])

	var icon string
	hasBuildResult := false

	switch status {
	case "START":
		{
			icon = ":white_medium_square:"
		}
	case "FAIL":
		{
			icon = ":exclamation:"
			hasBuildResult = true
		}
	case "SUCCESS":
		{
			icon = ":white_check_mark:"
			hasBuildResult = true
		}
	case "ABANDON":
		{
			icon = ":warning:"
			hasBuildResult = true
		}
	default:
		{
			icon = ":question:"
		}
	}

	var attachmentText string

	if hasBuildResult == true {
		attachmentText = fmt.Sprintf("%v%v - %v", icon, newformatted, buildResultUrl)
	} else {
		attachmentText = fmt.Sprintf("%v%v", icon, newformatted)
	}
	attachments := []Attachments{{
		Color:    color,
		Text:     attachmentText,
		Fallback: formatted,
	}}

	msg := SlackMsg{
		Channel:     conf.Channel,
		Username:    conf.Username,
		Attachments: attachments,
	}

	return msg
}

func doBuildRegexes(info map[string]string) ([]string, []string, []string) {
	var batchDeployRE = regexp.MustCompile(`Deploy to (.*) - DEPLOY ONE PROJECT.*`)
	var ssDeployRE = regexp.MustCompile(`(?i)Dev.* DEPLOY (.*?) - (.*)`)
	var ssBuildRE = regexp.MustCompile(`(?i)Dev.* SS.* Build - (.*)`)
	batchDeployRes := batchDeployRE.FindStringSubmatch(info["builddef"])
	ssDeployRes := ssDeployRE.FindStringSubmatch(info["builddef"])
	ssBuildRes := ssBuildRE.FindStringSubmatch(info["builddef"])
	return batchDeployRes, ssDeployRes, ssBuildRes
}

func getBuildInfo(buildStatus string, buildInfo BuildInfo, conf Configuration) SlackMsg {
	var info = buildInfo.Matches
	batchDeployRes, ssDeployRes, ssBuildRes := doBuildRegexes(info)
	if batchDeployRes != nil {
		return createBuildInfoMessage(buildInfo, buildStatus, batchDeployRes[1], info["projects"], conf)
	} else if ssDeployRes != nil {
		return createBuildInfoMessage(buildInfo, buildStatus, "SS - "+ssDeployRes[1], ssDeployRes[2], conf)
	} else if ssBuildRes != nil {
		var host string
		if conf.Hostname != "" {
			host = conf.Hostname
		} else {
			host = info["enghost"]
		}
		return createBuildInfoMessage(buildInfo, buildStatus, host, ssBuildRes[1], conf)
	} else {
		return createBuildInfoMessage(buildInfo, buildStatus, info["builddef"], info["projects"], conf)
	}
}

func formatBuildLogUrl(buildInfo BuildInfo, conf Configuration) string {
	var info = buildInfo.Matches
	return fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v", conf.RTCBaseURL, info["uuid"])
}
