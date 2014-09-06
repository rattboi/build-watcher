package main

import (
    "fmt"
    "encoding/json"
    "regexp"
    "flag"
    "log"
    "strings"
    "net"
    "os"
    "github.com/howeyc/fsnotify"
    "github.com/ActiveState/tail"
)

// Config stuff
type Configuration struct {
    Botname string
    Botaddress string
    Botport string
    Authpass string
    Watchdir string
    Filepattern string
    RTCBaseURL string
}


func setConfigDefaults(conf *Configuration) {
    conf.Botname     = "deploybot"
    conf.Botaddress  = "localhost"
    conf.Botport     = "12345" //default for ircflu
    conf.Authpass    = "password"
    conf.Watchdir    = "/tmp"
    conf.Filepattern = `.*build-(.*)\.log`
    conf.RTCBaseURL  = "http://baseurl:port/jazz"
}

func setupFlags(conf *Configuration) {
    flag.StringVar(&conf.Botname,    "Botname", conf.Botname, "irc bot name")
    flag.StringVar(&conf.Botaddress, "Botaddress", conf.Botaddress,"address where ircflu cat server is")
    flag.StringVar(&conf.Botport,    "Botport", conf.Botport, "port where ircflu cat server is")
    flag.StringVar(&conf.Authpass,   "Authpass", conf.Authpass, "password to auth the bot")
    flag.StringVar(&conf.Watchdir,   "Watchdir", conf.Watchdir, "directory to watch for build files")
    flag.StringVar(&conf.Filepattern,"Filepattern", conf.Filepattern, "regular expression pattern for files to watch")
    flag.StringVar(&conf.RTCBaseURL, "RTCBaseURL", conf.RTCBaseURL, "base part of RTC server for build log linking")
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

// Build Info Stuff
type BuildInfo struct {
    Patterns map[string] *regexp.Regexp
    Matches  map[string] string
}

func initBuildInfo(build *BuildInfo) {
    build.Patterns = make(map[string] *regexp.Regexp)
    build.Matches  = make(map[string] string)

    build.Patterns["uuid"], _      = regexp.Compile(`.*Build Result UUID: (.*)`)
    build.Patterns["requestor"], _ = regexp.Compile(`.*Requestor ID: (.*)`)
    build.Patterns["enghost"], _   = regexp.Compile(`.*Build Engine Host: (.*)`)
    build.Patterns["engid"], _     = regexp.Compile(`.*Build Engine ID: (.*)`)
    build.Patterns["builddef"], _  = regexp.Compile(`.*Build Definition: (.*)`)
    build.Patterns["projects"], _  = regexp.Compile(`.*Project Names: (.*)`)

    for k, _ := range build.Patterns {
        build.Matches[k] = ""
    }
}

// IRC Bot Helper
func WriteToIrcBot(message string, conf Configuration) {
    strEcho := message + "\n"
    servAddr := conf.Botaddress + ":" + conf.Botport
    tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)

    if err != nil {
        log.Println("ResolveTCPAddr failed:", err.Error())
    }

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    defer conn.Close()

    if err != nil {
        log.Println("Dial failed:", err.Error())
        return
    }

    _, err = conn.Write([]byte(strEcho))

    if err != nil {
        log.Println("Write to server failed:", err.Error())
        return
    }
}

// State Machine Stuff
const (
        Init = iota
        Start_Summ
        End_Summ
        Main_Log
        End_Log
        Exit
)

func Init_State(line string, conf Configuration, buildinfo BuildInfo) int {
    if strings.Contains(line, "-- START BUILD INFO --") { return Start_Summ } else { return Init }
}

func Start_Summ_State(line string, conf Configuration, buildinfo BuildInfo) int {

    var res[]string

    for k, v := range buildinfo.Patterns {
        res = v.FindStringSubmatch(line)
        if res != nil { buildinfo.Matches[k] = res[1] }
    }

    var allMatched bool = true
    for _, v := range buildinfo.Matches {
        if v == "" { allMatched = false }
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

func End_Summ_State(line string, conf Configuration, buildinfo BuildInfo) int {
    if strings.Contains(line, "-- END BUILD INFO --") { return Main_Log } else { return End_Summ }
}

func Main_Log_State(line string, conf Configuration, buildinfo BuildInfo) int {
    var builtLine string
    var info = buildinfo.Matches
    // Use HasPrefix instead of Contains in case of other exec'd ant jobs (like RTC tagging)
    if strings.HasPrefix(line, "BUILD SUCCESSFUL") {
        builtLine = fmt.Sprintf("SUCCESS: Requestor: %v, Project: %v, Def: %v", info["requestor"], info["projects"], info["builddef"])
        WriteToIrcBot(builtLine,conf)
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
    var builtLine string = fmt.Sprintf("%v/resource/itemOid/com.ibm.team.build.BuildResult/%v",conf.RTCBaseURL,build.Matches["uuid"])
    return builtLine
}

func End_Log_State(line string, conf Configuration, buildinfo BuildInfo) int {
    log.Println("Logfile finished")
    return Exit
}

func main() {
    var conf Configuration

    // set some dumb defaults
    setConfigDefaults(&conf)

    // get settings from config file if it exists and override defaults
    parseConfig(&conf)

    // parse out any flags and override defaults/config
    setupFlags(&conf)
    flag.Parse()

//    fmt.Printf("%+v\n", conf)
    states := map[int] func(string, Configuration, BuildInfo) int{
        Init: Init_State,
        Start_Summ: Start_Summ_State,
        End_Summ: End_Summ_State,
        Main_Log: Main_Log_State,
        End_Log: End_Log_State,
    }


    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    done := make(chan bool)

    // Process events
    go func() {
        for {
            select {
            case ev := <-watcher.Event:
                // See if a new file is created with build- in the name
                if strings.Contains(ev.Name,"build-") && ev.IsCreate() {
                    log.Println("New File -> ", ev.Name)
                    re, _ := regexp.Compile(conf.Filepattern)
                    res := re.FindStringSubmatch(ev.Name)
                    if res == nil {
                        log.Println("No match to " + conf.Filepattern)
                        return
                    }
                    buildid := res[1]
                    log.Println("Build ID -> ", buildid)

                    // 'fork' a new tail watcher for this build file
                    go func() {
                        t, err := tail.TailFile(ev.Name, tail.Config{Follow: true})
                        if err != nil {
                            return
                        }

                        var build BuildInfo
                        initBuildInfo(&build)

                        logState := Init
                        for line := range t.Lines {
                            logState = states[logState](line.Text, conf, build)
                            if (logState == Exit) { return }
                        }
                    }()
                }
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Watch(conf.Watchdir)
    if err != nil {
        log.Fatal(err)
    }

    <-done

    watcher.Close()
}
