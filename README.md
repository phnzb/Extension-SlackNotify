# SlackNotify

Simple NZBGet extension, written in Go

Post message in Slack when download finish.

# Building / Installing

- clone repository
- go build (for dynamic linking)
- CGO_ENABLED=0 go build (for static linking)
- copy manifest.json and slacknotify binary to NZBGet's {ScriptDir}/SlackNotify
- reload NZBGet, provide SlackToken and SlackChannel in settings, test
