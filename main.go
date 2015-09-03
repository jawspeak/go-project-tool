package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type Page struct {
	Team       string
	JiraStats  JiraStats
	StashStats StashStats
}

// JSON data type we return to manipulate for d3 visualization.
type JiraStats struct {
	Name           string
	CompletedCount int
}

type StashStats struct {
}

const JIRA_SERVER = "https://jira.corp.squareup.com"
const STASH_SERVER = "https://stash.corp.squareup.com"
const TEAM_NAME = "card-processing"

var TEAM_MEMBERS = [...]string{
	"botros",
	"davis",
	"dsimms",
	"jaw",
	"noam",
	"riley",
	"ryder",
}

func client() http.Client {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return http.Client{Transport: tr}
}

func callGet(server string, resource string, f func(http.Request)) http.Response {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", server, resource), nil)
	if f != nil {
		f(*req)
	}
	client := client()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("GET %s: %s\n", server, resp.Status)
	return *resp
}

func parseToJson(resp http.Response) map[string]interface{} {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}
	debugPrettyPrint(data)
	return data
}

func lookupFromJira() JiraStats {
	// query for stories unresolved per person
	parseToJson(callGet(JIRA_SERVER, "/rest/api/2/filter/18720",
		func(req http.Request) { req.SetBasicAuth("processing-bot", "n2nCmWF6") }))

	return JiraStats{
		Name:           "paul",
		CompletedCount: 10,
	}
}

func lookupFromStash() StashStats {
	// query for pull request contributions per person (created
	parseToJson(callGet(STASH_SERVER, "/rest/api/1.0/projects", nil))

	return StashStats{}
}

func debugPrettyPrint(data map[string]interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

func teamHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Team:       TEAM_NAME,
		JiraStats:  lookupFromJira(),
		StashStats: lookupFromStash(),
	}
	t, _ := template.ParseFiles("team.html")
	t.Execute(w, p)
}

func main() {
	fmt.Println("starting up...")
	http.HandleFunc(fmt.Sprintf("/team/%s", TEAM_NAME), teamHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
