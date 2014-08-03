package main

import (
    "flag"
    "log"
    "strings"
    "net"
    "os"
    "github.com/howeyc/fsnotify"
    "github.com/ActiveState/tail"
)

var botname string
var botaddress string
var botport string
var ircchan string
var authpass string
var watchdir string

func init() {
    flag.StringVar(&botname,    "ircbot", "ircflu", "irc bot name")
    flag.StringVar(&botaddress, "botaddress", "localhost", "address where ircflu cat server is")
    flag.StringVar(&botport,    "botport", "12345", "port where ircflu cat server is")
    flag.StringVar(&ircchan,    "ircchan", "", "channel to join after authing the bot")
    flag.StringVar(&authpass,   "authpass", "", "password to auth the bot")
    flag.StringVar(&watchdir,   "watchdir", "/tmp", "directory to watch for build files")

}

func WriteToIrcBot(message string) {
    strEcho := message + "\n"
    servAddr := botaddress + ":" + botport
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

func main() {
    // parse out flags
    flag.Parse()

    // a few arguments don't have defaults
    if ircchan == "" {
        println("Must supply ircchan argument")
        os.Exit(1);
    }

    if authpass == "" {
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
                    // 'fork' a new tail watcher for this build file
                    go func() {
                        t, err := tail.TailFile(ev.Name, tail.Config{Follow: true})
                        if err != nil {
                            return
                        }

                        WriteToIrcBot("@" + botname + " !auth " + authpass)
                        WriteToIrcBot("@" + botname + " !join #" + ircchan)

                        for line := range t.Lines {
                            WriteToIrcBot("#" + ircchan + " " + line.Text)
                            log.Println(line.Text)
                        }
                    }()
                }
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Watch(watchdir)
    if err != nil {
        log.Fatal(err)
    }

    <-done

    /* ... do stuff ... */
    watcher.Close()
}
