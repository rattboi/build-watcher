package buildwatcher

import (
    "regexp"
    "flag"
    "log"
    "strings"
    "net"
    "github.com/howeyc/fsnotify"
    "github.com/ActiveState/tail"
)

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
