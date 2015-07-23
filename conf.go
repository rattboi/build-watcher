package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

// Config stuff
type Configuration struct {
	Botaddress  string
	Botport     string
	Watchdir    string
	Filepattern string
	RTCBaseURL  string
	IRCChannel  string
}

func setConfigDefaults(conf *Configuration) {
	conf.Botaddress = "localhost"
	conf.Botport = "12345" //default for ircflu
	conf.Watchdir = "/tmp"
	conf.Filepattern = `.*build-(.*)\.log`
	conf.RTCBaseURL = "http://baseurl:port/jazz"
	conf.IRCChannel = "#channel"
}

func setupFlags(conf *Configuration) {
	flag.StringVar(&conf.Botaddress, "Botaddress", conf.Botaddress, "address where ircflu cat server is")
	flag.StringVar(&conf.Botport, "Botport", conf.Botport, "port where ircflu cat server is")
	flag.StringVar(&conf.Watchdir, "Watchdir", conf.Watchdir, "directory to watch for build files")
	flag.StringVar(&conf.Filepattern, "Filepattern", conf.Filepattern, "regular expression pattern for files to watch")
	flag.StringVar(&conf.RTCBaseURL, "RTCBaseURL", conf.RTCBaseURL, "base part of RTC server for build log linking")
	flag.StringVar(&conf.IRCChannel, "IRCChannel", conf.IRCChannel, "IRC channel to log to")
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
