package main

import (
	"html/template"
	"net/http"
)

func (app *application) htmxHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}