package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func Connexion() {
	http.HandleFunc("/profile_etudiant", profile_etudiant)

	http.HandleFunc("/login", login)
	err := http.ListenAndServe(":8080", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func profile_etudiant(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./web/html/profile_etudiant.html")
	if err != nil {
		fmt.Print("erreur chargement profile_etudiant.html")
	}
	t.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, err := template.ParseFiles("./web/html/accueil.html")
		if err != nil {
			fmt.Print("erreur chargement accueil.html")
		}
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
	}
}
