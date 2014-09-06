package main

import (
    "regexp"
)

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
