package main

import (
	"flag"
	"os"

	"github.com/jawspeak/go-project-tool/stashstats/data"
	"github.com/jawspeak/go-project-tool/stashstats/fetch"
	stashrestapiclientsetup "github.com/jawspeak/go-stash-restclient/config"
)

func main() {
	var mode = flag.String("mode", "", "[network|local] to use the local cache of data or re-fetch it")
	flag.Parse()

	var cache data.Cache
	switch *mode {
	case "":
		flag.Usage()
		os.Exit(1)
	case "remote":
		stashrestapiclientsetup.SetupConfig()
		cache = fetch.FetchData()
		// Fetching is probably slow, so store for playing with it later.
		cache.SaveGob("./pr-data-pull.dat")
		cache.SaveJson("./pr-data-pull.json")
	case "local":
		cache = data.LoadGob("./pr-data-pull.dat")
		cache.SaveJson("../assets/pr-data-pull.json")
	}
}
