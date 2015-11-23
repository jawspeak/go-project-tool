package fetch

import (
	"fmt"

	"log"

	"github.com/jawspeak/go-project-tool/stash-stats/data"
	"github.com/jawspeak/go-project-tool/stash-stats/util"
	apiclient "github.com/jawspeak/go-stash-restclient/client"
	"github.com/jawspeak/go-stash-restclient/client/operations"
	"github.com/jawspeak/go-stash-restclient/models"
)

type fetchConfig struct {
	Project string
	Repo    string
}

const (
	LIMIT int64 = 10
)

func FetchData() (cache data.Cache) {
	cache = data.Cache{}
	for _, toFetch := range []fetchConfig{fetchConfig{"GO", "square"}} {
		// TODO more projects and repos

		helper := LIMIT // ugly http://stackoverflow.com/questions/30716354/how-do-i-do-a-literal-int64-in-go
		pullRequests, err := apiclient.Default.Operations.GetPullRequests(
			operations.GetPullRequestsParams{Project: toFetch.Project, Repo: toFetch.Repo,
				Limit: &helper})
		util.FatalIfErr(err)
		fmt.Printf("%#v\n", pullRequests.Payload)
		for {
			log.Printf("fetched %d results", pullRequests.Payload.Size)

			for _, pr := range pullRequests.Payload.Values {
				activities, err := apiclient.Default.Operations.GetPullRequestActivities(
					operations.GetPullRequestActivitiesParams{
						Project:       toFetch.Project,
						Repo:          toFetch.Repo,
						PullRequestID: pr.ID})
				util.FatalIfErr(err)
				// Do some magic to make all the comments coalesce linearly.
				var accum []data.PrInteraction
				var commentsByAuthorLdap = make(map[string]int)
				var approvalsByAuthorLdap = make(map[string]int)
				closedAt := new(int64)
				for _, activity := range activities.Payload.Values {
					switch activity.Action {
					case "COMMENTED":
						flatten(&accum, activity.Comment, pr, commentsByAuthorLdap)
					case "OPENED":
						// ignore
					case "APPROVED":
						if closedAt == nil || activity.CreatedDatetime > *closedAt {
							*closedAt = activity.CreatedDatetime
						}
						if _, ok := approvalsByAuthorLdap[activity.User.Slug]; !ok {
							approvalsByAuthorLdap[activity.User.Slug] = 0
						}
						approvalsByAuthorLdap[activity.User.Slug] = approvalsByAuthorLdap[activity.User.Slug] + 1
						accum = append(accum, data.PrInteraction{
							Type:            "approval",
							RefId:           activity.ID,
							AuthorLdap:      activity.User.Slug,
							PullRequestId:   pr.ID,
							CreatedDateTime: activity.CreatedDatetime,
							PrApproval:      true})
						flatten(&accum, activity.Comment, pr, commentsByAuthorLdap)
					default:
						log.Printf("---> see other action state: %#v\n", activity)
					}
				}

				if *closedAt != pr.UpdatedDatetime {
					log.Printf("WARNING: pr closed at different time than last updated", pr.ID)
				}
				cache.PullRequests = append(cache.PullRequests, data.PullRequest{
					AuthorLdap:            pr.Author.User.Slug,
					Project:               toFetch.Project,
					Repo:                  toFetch.Repo,
					PullRequestId:         pr.ID,
					Title:                 pr.Title,
					CommentCount:          len(accum),
					Comments:              accum,
					CreatedDateTime:       pr.CreatedDatetime,
					UpdatedDateTime:       pr.UpdatedDatetime,
					SecondsOpen:           (pr.UpdatedDatetime - pr.CreatedDatetime),
					CommentsByAuthorLdap:  commentsByAuthorLdap,
					ApprovalsByAuthorLdap: approvalsByAuthorLdap,
				})
			}

			break // TODO loop while more activities. Now stop after 1 fetch
		}
	}
	return cache
}

func flatten(accum *[]data.PrInteraction, input *models.Comment, contextPr *models.PullRequest,
	commentsByAuthorLdap map[string]int) {
	if input == nil {
		return // skip empty comments (e.g. in Approved activities
	}
	if _, ok := commentsByAuthorLdap[input.Author.Slug]; !ok {
		commentsByAuthorLdap[input.Author.Slug] = 0
	}
	commentsByAuthorLdap[input.Author.Slug] = commentsByAuthorLdap[input.Author.Slug] + 1
	*accum = append(*accum, data.PrInteraction{
		Type:            "comment",
		RefId:           input.ID,
		AuthorLdap:      input.Author.Slug,
		PullRequestId:   contextPr.ID,
		CreatedDateTime: input.CreatedDatetime,
		PrApproval:      false, // comments don't have approvals
	})
	for _, nested := range input.Comments {
		flatten(accum, nested, contextPr, commentsByAuthorLdap)
	}
}
