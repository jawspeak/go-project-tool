## Simple Stash stats team server

1. `go build ./...` from the root directory
1. `cd stashstats` and create your own `config.json` based on the samele. Then with the instructions
 `go run main.go -mode=...` to download and cache your stash stats.
1. Back in the root directory, `go run main.go` to start the webserver, then browse.

Productionizing:

* Run the stashstats on a cron job. This updates the cached json file for each team member.
