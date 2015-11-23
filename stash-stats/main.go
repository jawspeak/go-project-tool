package main

import (
	"fmt"
	"log"

	apiclient "github.com/jawspeak/go-stash-restclient/client"
	"github.com/jawspeak/go-stash-restclient/client/operations"
	apiclientsetup "github.com/jawspeak/go-stash-restclient/config"
)

func main() {
	apiclientsetup.SetupConfig()

	resp2, err := apiclient.Default.Operations.GetPullRequests(operations.GetPullRequestsParams{Project: "GO", Repo: "square"})
	fatalIfErr(err)
	fmt.Printf("%#v\n", resp2.Payload)
	pullRequestId := resp2.Payload.Values[0].ID

	resp3, err := apiclient.Default.Operations.GetPullRequestActivities(
		operations.GetPullRequestActivitiesParams{Project: "GO", Repo: "square", PullRequestID: pullRequestId})
	fatalIfErr(err)
	fmt.Printf("%#v\n", resp3.Payload)
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
