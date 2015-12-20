package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Page struct {
	Title string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{Title: "Developers Stash Dashboard"}
	t, _ := template.ParseFiles("templates/index.thtml")
	t.Execute(w, p)
}

func doNotCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
		h.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("started up!")
	http.Handle("/assets/", doNotCache(http.StripPrefix("/assets", http.FileServer(http.Dir("assets")))))
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
