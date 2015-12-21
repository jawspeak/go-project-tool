package main

import (
	"flag"
	"os"

	"github.com/jawspeak/go-project-tool/stashstats/data"
	"github.com/jawspeak/go-project-tool/stashstats/fetch"
	stashrestapiclientsetup "github.com/jawspeak/go-stash-restclient/config"
)

func main() {
	var mode = flag.String("mode", "", "[remote|local] to use the local cache of data or re-fetch it")
	flag.Parse()

	var cache data.Cache
	switch *mode {
	case "remote":
		stashrestapiclientsetup.SetupConfig()
		cache = fetch.FetchData()
		// Fetching is slow, so store for playing with it later.
		cache.SaveGob("./pr-data-pull.dat")
		cache.SaveJson("../assets/pr-data-pull.json")
	case "local":
		cache = data.LoadGob("./pr-data-pull.dat")
		cache.SaveJson("../assets/pr-data-pull.json")
	case "":
		fallthrough
	default:
		flag.Usage()
		os.Exit(1)
	}
}
