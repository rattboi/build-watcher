package main

import (
    "log"
    "strings"
    "net"
    "os"
    "github.com/howeyc/fsnotify"
    "github.com/ActiveState/tail"
)

func main() {
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
                    log.Println("file created. Name -> ", ev.Name)
                    // 'fork' a new tail watcher for this build file
                    go func() {
                        t, err := tail.TailFile(ev.Name, tail.Config{Follow: true})
                        if err != nil {
                            return
                        }
                        for line := range t.Lines {
                            strEcho := line.Text + "\n"
                            servAddr := "localhost:12345"
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
                                os.Exit(1)
                            }

                            println("write to server = ", strEcho)

                            conn.Close()

                            log.Println(line.Text)
                        }
                    }()
                }
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    // Watching /tmp for now. Need to parameterize
    err = watcher.Watch("/tmp")
    if err != nil {
        log.Fatal(err)
    }

    <-done

    /* ... do stuff ... */
    watcher.Close()
}
