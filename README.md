Build-watcher
==============

First golang project

1. Watch for new files in a log directory
2. Tail each new file via a goroutine
3. Write out log events as Incoming Webhook requests to Slack

This is a component of a system to forward build/deploy logs summaries to Slack
