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

// JSON data type we return to populate d3.
type Jira struct {
	Name           string
	CompletedCount int
}

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

func lookupFromJira() Jira {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest("GET", "https://jira.corp.squareup.com/rest/api/2/filter/18720", nil)
	req.SetBasicAuth("processing-bot", "n2nCmWF6")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)
	fmt.Println(resp.Header)

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

	prettyPrint(data)
	fmt.Printf("id is %s", data["id"])

	return Jira{
		Name:           "paul",
		CompletedCount: 10,
	}
}

func prettyPrint(data map[string]interface{}) {
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
