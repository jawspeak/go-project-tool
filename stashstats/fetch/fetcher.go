package fetch

import (
	"fmt"

	"log"

	"math/rand"
	"sync"
	"time"

	"github.com/go-swagger/go-swagger/vendor/github.com/davecgh/go-spew/spew"
	"github.com/jawspeak/go-project-tool/stashstats/data"
	"github.com/jawspeak/go-project-tool/stashstats/util"
	apiclient "github.com/jawspeak/go-stash-restclient/client"
	"github.com/jawspeak/go-stash-restclient/client/operations"
	"github.com/jawspeak/go-stash-restclient/config"
	"github.com/jawspeak/go-stash-restclient/models"
)

type fetchConfig struct {
	LookBackDays           int             `json:"look_back_days"`
	Projects               []projectConfig `json:"stash_projects"`
	Usernames              []string        `json:"stats_for_these_usernames"`
	IgnoreCommentUsernames []string        `json:"ignore_comments_from_usernames"`
}
type projectConfig struct {
	Project string   `json:"project"`
	Repos   []string `json:"repos"`
}

// passing into a work channel which is rate limited.
type fetchOneWork struct {
	project              string
	repo                 string
	author               string
	resultChan           chan data.PullRequest
	wg                   *sync.WaitGroup
	lookBackUntil        int64
	ignoreCommentAuthors []string
}

const (
	LIMIT           int64 = 50 // How many to fetch at a time
	MAX_CONCURRENCY       = 30
)

func FetchData() (cache data.Cache) {
	spew.Config.MaxDepth = 1

	var conf fetchConfig
	config.ParseJsonFileStripComments("./config.json", &conf)
	fmt.Printf("Processing for config: %+v\n", conf)

	var wg sync.WaitGroup
	workChan := make(chan fetchOneWork)
	resultChan := make(chan data.PullRequest)

	// Many goroutines to enqueue up results to do work.
	for i := 0; i < MAX_CONCURRENCY; i++ {
		go func() {
			for work := range workChan {
				fmt.Println("got some work!")
				spew.Dump(work)
				fetchOne(&work)
				fmt.Printf("work done\n")
			}
		}()
	}
	cache = data.Cache{}
	// One goroutine to work on the results. Start it now so it can start working.
	go func() {
		for result := range resultChan {
			cache.PullRequests = append(cache.PullRequests, result)
			wg.Done()
		}
	}()

	lookBackUntil := time.Now().Unix() - int64(conf.LookBackDays*24*60*60)
	fmt.Println("All workers are started, looking back until: ", time.Unix(lookBackUntil, 0))

	var workEnqueed []string
	for _, confProjects := range conf.Projects {
		for _, confAuthor := range conf.Usernames {
			for _, confRepo := range confProjects.Repos {
				workChan <- fetchOneWork{
					project:              confProjects.Project,
					repo:                 confRepo,
					author:               confAuthor,
					resultChan:           resultChan,
					wg:                   &wg,
					lookBackUntil:        lookBackUntil,
					ignoreCommentAuthors: conf.IgnoreCommentUsernames}
				workEnqueed = append(workEnqueed, fmt.Sprintf("%v-%v-%v", confProjects.Project, confAuthor, confRepo))
				beNiceFuzzySleep()
			}
		}
	}

	fmt.Println("All work is enqueued", workEnqueed)
	// close the channel when all are done
	wg.Wait()
	fmt.Println("done waiting")
	close(workChan) // ok to close both
	close(resultChan)
	fmt.Println("channels closed")

	// sort pull requests by author or something else if i cared, but i don't.
	return cache
}

// Fetch all the data!
func fetchOne(work *fetchOneWork) {
	defer func() {
		beNiceFuzzySleep()
		work.wg.Done() // Remove 1 we added at the top.
	}()

	work.wg.Add(1) // Add 1 at the top, and pop that when done in defer. Also add one every time we push to the resultChan.
	start := int64(0)
	for {
		limitHelper := LIMIT // ugly http://stackoverflow.com/questions/30716354/how-do-i-do-a-literal-int64-in-go
		orderHelper := "NEWEST"
		role1Helper := "AUTHOR"
		stateHelper := "ALL"
		prParams := operations.GetPullRequestsParams{
			Project:   work.project,
			Repo:      work.repo,
			Username1: &work.author,
			Role1:     &role1Helper,
			Limit:     &limitHelper,
			Order:     &orderHelper,
			Start:     &start,
			State:     &stateHelper}
		fmt.Printf("> %s\n", spew.Sdump(prParams))
		pullRequests, err := apiclient.Default.Operations.GetPullRequests(prParams)
		if util.FatalIfErrUnless(err, okIf404, prParams) {
			continue // skip if error
		}
		fmt.Printf("< %s\n", spew.Sdump(pullRequests.Payload))
		log.Printf("fetched %d results", pullRequests.Payload.Size)

		start = pullRequests.Payload.NextPageStart

		for _, pr := range pullRequests.Payload.Values {
			if msToSec(pr.CreatedDate) < work.lookBackUntil {
				fmt.Println("Won't look back any further, skipping", pr.Author.User.Slug,
					"'s PR ", pr.ID, " at: ",
					time.Unix(msToSec(pr.CreatedDate), 0))
				break
			}

			actParams := operations.GetPullRequestActivitiesParams{
				Project:       work.project,
				Repo:          work.repo,
				PullRequestID: pr.ID,
				Limit:         &limitHelper}
			fmt.Printf(">> %s\n", spew.Sdump(actParams))
			activities, err := apiclient.Default.Operations.GetPullRequestActivities(actParams)
			if util.FatalIfErrUnless(err, okIf404, actParams) {
				continue
			}
			fmt.Printf("<< %s\n", spew.Sdump(activities))

			// Do some magic to make all the comments coalesce linearly.
			var accum []data.PrInteraction
			var commentsByAuthorLdap = make(map[string]int)
			var approvalsByAuthorLdap = make(map[string]int)
			approvedAt := new(int64)
			for _, activity := range activities.Payload.Values {
				switch activity.Action {
				case "COMMENTED":
					flatten(work.ignoreCommentAuthors, &accum, activity.Comment, pr,
						commentsByAuthorLdap)
				case "OPENED":
				// ignore
				case "APPROVED":
					if approvedAt == nil || msToSec(activity.CreatedDate) > *approvedAt {
						*approvedAt = msToSec(activity.CreatedDate)
					}
					if contains(work.ignoreCommentAuthors, activity.User.Slug) {
						continue // Skip this activity.
					}

					// Mark the approval.
					if _, ok := approvalsByAuthorLdap[activity.User.Slug]; !ok {
						approvalsByAuthorLdap[activity.User.Slug] = 0
					}
					approvalsByAuthorLdap[activity.User.Slug] = approvalsByAuthorLdap[activity.User.Slug] + 1

					// Mark the approval as a comment, too.
					if _, ok := commentsByAuthorLdap[activity.User.Slug]; !ok {
						commentsByAuthorLdap[activity.User.Slug] = 0
					}
					commentsByAuthorLdap[activity.User.Slug] = commentsByAuthorLdap[activity.User.Slug] + 1

					accum = append(accum, data.PrInteraction{
						Type:            "approval",
						RefId:           activity.ID,
						AuthorLdap:      activity.User.Slug,
						AuthorFullName:  activity.User.DisplayName,
						PullRequestId:   pr.ID,
						CreatedDateTime: msToSec(activity.CreatedDate),
						PrApproval:      true})
					flatten(work.ignoreCommentAuthors, &accum, activity.Comment, pr,
						commentsByAuthorLdap)
				case "RESCOPED":
					// adding or removing of commits. don't care. ignore.
				case "MERGED":
					// ignore
				case "DECLINED":
					// ignore
				case "UNAPPROVED":
					// ignore
				default:
					log.Printf("---> see %s other action state: %#v\n", pr.ID, activity)
				}
			}
			// Don't care, but if I did here are interesting records.
			//						if *approvedAt != msToSec(pr.UpdatedDate) {
			//							log.Printf("Note: %s approvedAt=%s different than last updated %s",
			//								pr.ID, *approvedAt, msToSec(pr.UpdatedDate))
			//						}

			fmt.Println("to push to resultChan")
			work.resultChan <- data.PullRequest{
				AuthorLdap:            pr.Author.User.Slug,
				AuthorFullName:        pr.Author.User.DisplayName,
				Project:               work.project,
				Repo:                  work.repo,
				PullRequestId:         pr.ID,
				Title:                 pr.Title,
				CommentCount:          len(accum),
				Comments:              accum,
				CreatedDateTime:       msToSec(pr.CreatedDate),
				UpdatedDateTime:       msToSec(pr.UpdatedDate),
				SecondsOpen:           msToSec(pr.UpdatedDate - pr.CreatedDate),
				CommentsByAuthorLdap:  commentsByAuthorLdap,
				ApprovalsByAuthorLdap: approvalsByAuthorLdap,
				State:   pr.State,
				SelfUrl: pr.Links.Self[0].Href,
			}
			work.wg.Add(1)
			fmt.Println("pushed to resultChan")
		}
		if pullRequests.Payload.IsLastPage {
			break
		}
	}
}

func beNiceFuzzySleep() {
	time.Sleep(time.Duration(rand.Float32()+1) * time.Second)
}

func msToSec(ms int64) int64 {
	return ms / 1000
}

func flatten(ignoreCommentAuthors []string, accum *[]data.PrInteraction, input *models.Comment,
	contextPr *models.PullRequest, commentsByAuthorLdap map[string]int) {
	if input == nil {
		return // skip empty comments (e.g. in Approved activities
	}
	if contains(ignoreCommentAuthors, input.Author.Slug) {
		return // ignore comments, and comment threads started by these authors
	}
	if _, ok := commentsByAuthorLdap[input.Author.Slug]; !ok {
		commentsByAuthorLdap[input.Author.Slug] = 0
	}
	commentsByAuthorLdap[input.Author.Slug] = commentsByAuthorLdap[input.Author.Slug] + 1
	*accum = append(*accum, data.PrInteraction{
		Type:            "comment",
		RefId:           input.ID,
		AuthorLdap:      input.Author.Slug,
		AuthorFullName:  input.Author.DisplayName,
		PullRequestId:   contextPr.ID,
		CreatedDateTime: msToSec(input.CreatedDate),
		PrApproval:      false, // comments don't have approvals
	})
	for _, nested := range input.Comments {
		flatten(ignoreCommentAuthors, accum, nested, contextPr, commentsByAuthorLdap)
	}
}

func okIf404(err error) bool {
	if apiErr, ok := err.(operations.APIError); ok {
		if apiErr.Code == 404 {
			fmt.Println("404 Not Found - skipping", apiErr)
			return true
		}
	}
	return false
}

func contains(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}
