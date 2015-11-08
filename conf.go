package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

// Config stuff
type Configuration struct {
	WebhookUrl  string
	Username    string
	Channel     string
	Watchdir    string
	Filepattern string
	RTCBaseURL  string
}

func setConfigDefaults(conf *Configuration) {
	conf.WebhookUrl = "localhost"
	conf.Username = "watchbot"
	conf.Channel = ""
	conf.Watchdir = "/tmp"
	conf.Filepattern = `.*build-(.*)\.log`
	conf.RTCBaseURL = "http://baseurl:port/jazz"
}

func setupFlags(conf *Configuration) {
	flag.StringVar(&conf.WebhookUrl, "WebhookUrl", conf.WebhookUrl, "Web Hook URL for Slack")
	flag.StringVar(&conf.Username, "Username", conf.Username, "Username to display in Slack")
	flag.StringVar(&conf.Channel, "Channel", conf.Channel, "Slack Channel to log to")
	flag.StringVar(&conf.Watchdir, "Watchdir", conf.Watchdir, "directory to watch for build files")
	flag.StringVar(&conf.Filepattern, "Filepattern", conf.Filepattern, "regular expression pattern for files to watch")
	flag.StringVar(&conf.RTCBaseURL, "RTCBaseURL", conf.RTCBaseURL, "base part of RTC server for build log linking")
	flag.Parse()
}

func parseConfig(conf *Configuration) {
	file, err := os.Open("conf.json")
	if err != nil {
		// assumes no config file. Not an error. Probably should further parse to verify...
		return
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}
}

func handleConfig() Configuration {
	var conf Configuration
	// set some dumb defaults
	setConfigDefaults(&conf)
	// get settings from config file if it exists and override defaults
	parseConfig(&conf)
	// parse out any flags and override defaults/config
	setupFlags(&conf)
	return conf
}
