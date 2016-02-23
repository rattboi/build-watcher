Build-watcher
==============

Forwards build/deploy start/success/fail notifications from Jazz Build Engines to Slack 

Operation
=========

1. Watch for new files in a log directory
2. Tail each new file via a goroutine
3. Write out log events as Incoming Webhook requests to Slack

Build
=====

1. `go get github.com/rattboi/build-watcher`
1. `make`
1. configure conf.json
1. `$GOBIN/build-watcher`

Configuration
=============

1. WebhookUrl: Slack webhook 
1. Username: Bot's username
1. Channel: Channel to publish slack messages to 
1. Watchdir: Directory to watch for log files
1. Filepattern: Regular expression for files to match and follow
1. RTCBaseURL: Base URL to RTC, to give build results links at completion of a build job

Screenshots
===========

Slack - Build Summary:

![alt text](https://github.com/rattboi/build-watcher/raw/master/docs/screenshot-build.png "Slack - Build Screenshot")

Slack - Deploy Summary:

![alt text](https://github.com/rattboi/build-watcher/raw/master/docs/screenshot-deploy.png "Slack - Deploy Screenshot")

