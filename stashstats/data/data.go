package data

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"

	"github.com/jawspeak/go-project-tool/stashstats/util"
)

type Cache struct {
	PullRequests []PullRequest `json:"pull_requests"`
}

type PullRequest struct {
	AuthorLdap     string `json:"author_ldap"`
	AuthorFullName string `json:"author_fullname"`
	Project        string `json:"project"`
	Repo           string `json:"repo"`
	PullRequestId  int64  `json:"pull_request_id"`
	Title          string `json:"title"`
	CommentCount   int    `json:"comment_count"`
	/* e.g. OPEN, MERGED, REJECTED */
	State                 string          `json:"state"`
	Comments              []PrInteraction `json:"comments"`
	CreatedDateTime       int64           `json:"created_datetime"`
	UpdatedDateTime       int64           `json:"updated_datetime"`
	SecondsOpen           int64           `json:"seconds_open"`
	CommentsByAuthorLdap  map[string]int  `json:"comments_by_author_ldap"`
	ApprovalsByAuthorLdap map[string]int  `json:"approvals_by_author_ldap"`
	SelfUrl               string          `json:"self_url"`
}

// A comment or approval
type PrInteraction struct {
	AuthorLdap      string `json:"author_ldap"`
	AuthorFullName  string `json:"author_fullname"`
	PullRequestId   int64  `json:"pull_request_id"`
	CreatedDateTime int64  `json:"created_datetime"`
	PrApproval      bool   `json:"approved"`
	Type            string `json:"type"`
	RefId           int64  `json:"ref_id"`
}

func (cache *Cache) SaveGob(filepath string) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(cache)
	util.FatalIfErr(err)
	err = ioutil.WriteFile(filepath, b.Bytes(), 0644)
	util.FatalIfErr(err)
}

func LoadGob(filepath string) (cache Cache) {
	cache = Cache{}
	file, err := ioutil.ReadFile(filepath)
	b := bytes.Buffer{}
	b.Write(file)
	err = gob.NewDecoder(&b).Decode(&cache)
	util.FatalIfErr(err)
	return cache
}

func (cache *Cache) SaveJson(filepath string) {
	dat, err := json.Marshal(cache)
	util.FatalIfErr(err)
	ioutil.WriteFile(filepath, dat, 0644)
}
