package main

import (
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

var conf Configuration

func setupFlags() {
    flag.StringVar(&conf.Botname,    "Botname", conf.Botname, "irc bot name")
    flag.StringVar(&conf.Botaddress, "Botaddress", conf.Botaddress,"address where ircflu cat server is")
    flag.StringVar(&conf.Botport,    "Botport", conf.Botport, "port where ircflu cat server is")
    flag.StringVar(&conf.Chanprefix, "Chanprefix", conf.Chanprefix, "channel to join after authing the bot")
    flag.StringVar(&conf.Authpass,   "Authpass", conf.Authpass, "password to auth the bot")
    flag.StringVar(&conf.Watchdir,   "Watchdir", conf.Watchdir, "directory to watch for build files")
}

func WriteToIrcBot(message string) {
    strEcho := message + "\n"
    servAddr := conf.Botaddress + ":" + conf.Botport
    tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)

    if err != nil {
        println("ResolveTCPAddr failed:", err.Error())
        os.Exit(1)
    }

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        println("Dial failed:", err.Error())
        os.Exit(1)
    }

    _, err = conn.Write([]byte(strEcho))

    if err != nil {
        println("Write to server failed:", err.Error())
        return
    }
    conn.Close()
}

func parseConfig() {
    file, _ := os.Open("conf.json")
    decoder := json.NewDecoder(file)

    err := decoder.Decode(&conf)
    if err != nil {
       println("error:", err)
    }
}

func main() {
    // get settings from config file
    parseConfig()

    // parse out any flags and override
    setupFlags()
    flag.Parse()

    // a few arguments don't have defaults
    if conf.Chanprefix == "" {
        println("Must supply Chanprefix argument")
        os.Exit(1);
    }

    if conf.Authpass == "" {
        println("Must supply authpass argument")
        os.Exit(1);
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

                        WriteToIrcBot("@" + conf.Botname + " !auth " + conf.Authpass)
                        WriteToIrcBot("@" + conf.Botname + " !join " + buildchan)

                        WriteToIrcBot("Work started - follow in channel: " + buildchan)

                        for line := range t.Lines {
                            WriteToIrcBot(buildchan + " " + line.Text)
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

    /* ... do stuff ... */
    watcher.Close()
}
