package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"crypto/tls"
	"strings"
	"encoding/json"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
)

type Page struct {
	Team string
}

func getJiraData() {
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
	
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)

	fmt.Println(jq)
}

// getStashData

func teamHandler(w http.ResponseWriter, r *http.Request) {
	getJiraData()
	p := &Page{Team: "card processing"}
	t, _ := template.ParseFiles("team.html")
	t.Execute(w, p)
}

func main() {
	fmt.Println("hi")

	http.HandleFunc("/team/card-processing", teamHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
