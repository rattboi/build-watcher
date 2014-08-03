package main

import (
    "log"
    "strings"
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
                            if err == nil {
                                for line := range t.Lines {
                                        log.Println(line.Text)
                                }
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
