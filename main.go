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
	Team string
	Jira Jira // json format
}

// JSON data type we return to manipulate for d3 visualization.
type Jira struct {
	Name           string
	CompletedCount int
}

const JIRA_SERVER = "https://jira.corp.squareup.com"
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

func callJiraGet(resource string) http.Response {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", JIRA_SERVER, resource), nil)
	req.SetBasicAuth("processing-bot", "n2nCmWF6")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.Status)
	log.Println(resp.Header)
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

func lookupFromJira() Jira {
	// query for stories unresolved per person
	resp := callJiraGet("/rest/api/2/filter/18720") // TODO new filter
	parseToJson(resp)
	//fmt.Printf("id is %s", data["id"])

	return Jira{
		Name:           "paul",
		CompletedCount: 10,
	}
}

func debugPrettyPrint(data map[string]interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

// TODO lookupFromStash

func teamHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Team: TEAM_NAME,
		Jira: lookupFromJira(),
	}
	t, _ := template.ParseFiles("team.html")
	t.Execute(w, p)
}

func main() {
	fmt.Println("hi")
	fmt.Println(fmt.Sprintf("/team/%s", TEAM_NAME))
	http.HandleFunc(fmt.Sprintf("/team/%s", TEAM_NAME), teamHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
