package main

import (
//    "fmt"
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

type Configuration struct {
    Botname string
    Botaddress string
    Botport string
    Chanprefix string
    Authpass string
    Watchdir string
}

func setConfigDefaults(conf *Configuration) {
    conf.Botname = "deploybot"
    conf.Botaddress = "localhost"
    conf.Botport = "12345" //default for ircflu
    conf.Chanprefix = "#work-"
    conf.Authpass = "password"
    conf.Watchdir = "/tmp"
}

func setupFlags(conf *Configuration) {
    flag.StringVar(&conf.Botname,    "Botname", conf.Botname, "irc bot name")
    flag.StringVar(&conf.Botaddress, "Botaddress", conf.Botaddress,"address where ircflu cat server is")
    flag.StringVar(&conf.Botport,    "Botport", conf.Botport, "port where ircflu cat server is")
    flag.StringVar(&conf.Chanprefix, "Chanprefix", conf.Chanprefix, "channel to join after authing the bot")
    flag.StringVar(&conf.Authpass,   "Authpass", conf.Authpass, "password to auth the bot")
    flag.StringVar(&conf.Watchdir,   "Watchdir", conf.Watchdir, "directory to watch for build files")
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
                    re, _ := regexp.Compile(`.*build-(.*)\.txt`)
                    res := re.FindStringSubmatch(ev.Name)
                    if res == nil {
                        log.Println("No match to build-*.txt")
                        return
                    }
                    buildid := res[1]
                    log.Println("Build ID -> ", buildid)
                    buildchan := conf.Chanprefix + buildid
                    log.Println("Creating channel at -> ", buildchan)

                    // 'fork' a new tail watcher for this build file
                    go func() {
                        t, err := tail.TailFile(ev.Name, tail.Config{Follow: true})
                        if err != nil {
                            return
                        }

                        WriteToIrcBot("@" + conf.Botname + " !auth " + conf.Authpass, conf)
                        WriteToIrcBot("@" + conf.Botname + " !join " + buildchan, conf)

                        WriteToIrcBot("Work started - follow in channel: " + buildchan, conf)

                        for line := range t.Lines {
                            WriteToIrcBot(buildchan + " " + line.Text, conf)
                            log.Println(line.Text)
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
