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
    Authpass string
    Watchdir string
    Filepattern string
}

func setConfigDefaults(conf *Configuration) {
    conf.Botname = "deploybot"
    conf.Botaddress = "localhost"
    conf.Botport = "12345" //default for ircflu
    conf.Authpass = "password"
    conf.Watchdir = "/tmp"
    conf.Filepattern = `.*build-(.*)\.log`
}

func setupFlags(conf *Configuration) {
    flag.StringVar(&conf.Botname,    "Botname", conf.Botname, "irc bot name")
    flag.StringVar(&conf.Botaddress, "Botaddress", conf.Botaddress,"address where ircflu cat server is")
    flag.StringVar(&conf.Botport,    "Botport", conf.Botport, "port where ircflu cat server is")
    flag.StringVar(&conf.Authpass,   "Authpass", conf.Authpass, "password to auth the bot")
    flag.StringVar(&conf.Watchdir,   "Watchdir", conf.Watchdir, "directory to watch for build files")
    flag.StringVar(&conf.Filepattern,"Filepattern", conf.Filepattern, "regular expression pattern for files to watch")
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

                        WriteToIrcBot("Work started - build: " + buildid , conf)

                        uuid_re, _ := regexp.Compile(`.*Build Result UUID: (.*)`)
                        requestor_re, _ := regexp.Compile(`.*Requestor ID: (.*)`)
                        enghost_re, _ := regexp.Compile(`.*Build Engine Host: (.*)`)
                        engid_re, _ := regexp.Compile(`.*Build Engine ID: (.*)`)
                        builddef_re, _ := regexp.Compile(`.*Build Definition: (.*)`)
                        projects_re, _ := regexp.Compile(`.*Project Names: (.*)`)

                        uuid, requestor, enghost, engid, builddef, projects := "","","","","",""

                        summary_block := false
                        for line := range t.Lines {
                            if strings.Contains(line.Text, "-- START BUILD INFO --") { summary_block = true  }

                            if summary_block == true {
                                res := uuid_re.FindStringSubmatch(line.Text)
                                if res != nil { uuid = res[1] }
                                res = requestor_re.FindStringSubmatch(line.Text)
                                if res != nil { requestor = res[1] }
                                res = enghost_re.FindStringSubmatch(line.Text)
                                if res != nil { enghost = res[1] }
                                res = engid_re.FindStringSubmatch(line.Text)
                                if res != nil { engid = res[1] }
                                res = builddef_re.FindStringSubmatch(line.Text)
                                if res != nil { builddef = res[1] }
                                res = projects_re.FindStringSubmatch(line.Text)
                                if res != nil { projects = res[1] }

                                log.Println(buildid + ": " + line.Text)
                            }
                            if uuid != "" && requestor != "" && enghost != "" && engid != "" && builddef != "" && projects != "" && summary_block == true {
                                WriteToIrcBot("UUID: " + uuid + ",    Req: " + requestor + ",     Eng: " + enghost + ":" + engid + ",     Def: " + builddef + ",    Project: " + projects, conf)
                                summary_block = false
                            }

                            if strings.Contains(line.Text, "-- END BUILD INFO --")   { summary_block = false }
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
