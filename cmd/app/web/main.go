package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
)

func main() {

	addr := flag.String("addr", ":8080", "Server address")

	flag.Parse()
	http.HandleFunc("/", index)
	http.HandleFunc("/preflight", preflight)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./cmd/app/web/pages/index.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		return
	}
}

func preflight(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./cmd/app/web/pages/preflight.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		return
	}
}
