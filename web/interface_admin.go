package web

import (
	"html/template"
	"log"
	"net/http"
)

type Admin struct {
}

func pageAdmin(w http.ResponseWriter, r *http.Request) {

	t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))

	if err := t.Execute(w, nil); err != nil {
		log.Printf("error exec template : ", err)
	}
}
