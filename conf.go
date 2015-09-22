package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

// Config stuff
type Configuration struct {
	Redishost   string
	Watchdir    string
	Filepattern string
}

func setConfigDefaults(conf *Configuration) {
	conf.Redishost = "localhost"
	conf.Watchdir = "/tmp"
	conf.Filepattern = `.*build-(.*)\.log`
}

func setupFlags(conf *Configuration) {
	flag.StringVar(&conf.Redishost, "Redishost", conf.Redishost, "host of redis queue")
	flag.StringVar(&conf.Watchdir, "Watchdir", conf.Watchdir, "directory to watch for build logs")
	flag.StringVar(&conf.Filepattern, "Filepattern", conf.Filepattern, "regular expression pattern for files to watch")
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
