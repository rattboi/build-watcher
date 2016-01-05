Build-watcher
==============

First golang project

1. Watch for new files in a log directory
2. Tail each new file via a goroutine
3. Write out log events as Incoming Webhook requests to Slack

This is a component of a system to forward build/deploy logs summaries to Slack.

Slack - Build Summary:

![alt text](https://github.com/rattboi/build-watcher/raw/master/screenshot-build.png "Slack - Build Screenshot")

Slack - Deploy Summary:

![alt text](https://github.com/rattboi/build-watcher/raw/master/screenshot-deploy.png "Slack - Deploy Screenshot")
