package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"html/template"
	"log"
	"net/http"
)

func InitWeb() {

	http.HandleFunc("/login", accueil)
	http.HandleFunc("/pageEtudiant", pageEtudiant)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "POST" {

		login := r.FormValue("login")
		password := r.FormValue("password")

		etu := BDD.LoginCorrect(login, password)
		fmt.Println(etu)
		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))

		err := t.Execute(w, etu)
		if err != nil {
			log.Printf("error exec template : ", err)
		}
	}
}
func accueil(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		t, err := template.ParseFiles("./web/html/accueil.html")
		if err != nil {
			fmt.Print("erreur chargement accueil.html")
		}
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in
		fmt.Println("oui")

	}

}
func upload() {

}
