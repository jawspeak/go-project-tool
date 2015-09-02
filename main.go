package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Page struct {
	Team string
}

func teamHandler(w http.ResponseWriter, r *http.Request) {

	p := &Page{Team: "card processing"}
	t, _ := template.ParseFiles("team.html")
	t.Execute(w, p)
}

func main() {
	fmt.Println("hi")

	http.HandleFunc("/team/card-processing", teamHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("bye")
}
