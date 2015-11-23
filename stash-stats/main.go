package main

import (
	"github.com/jawspeak/go-project-tool/stash-stats/data"
	"github.com/jawspeak/go-project-tool/stash-stats/fetch"
	stashrestapiclientsetup "github.com/jawspeak/go-stash-restclient/config"
)

func main() {
	stashrestapiclientsetup.SetupConfig()

	var cache data.Cache
	cache = fetch.FetchData()

	// Fetching is probably slow, so store for playing with it later.
	// cache = data.LoadGob("./pr-data-pull.dat")

	cache.SaveGob("./pr-data-pull.dat")

	cache.SaveJson("./pr-data-pull.json")
}
