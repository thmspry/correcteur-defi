package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"html/template"
	"log"
	"net/http"
)

var etudiantCo BDD.Etudiant

func InitWeb() {

	http.HandleFunc("/login", accueil) // Page d'acceuil : http://localhost:8080/login

	http.HandleFunc("/pageEtudiant", pageEtudiant) // Page étudiant : http://localhost:8080/pageEtudiant
	err := http.ListenAndServe(":8080", nil)       // port utilisé
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de pageEtudiant :", r.Method)
	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {

		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))

		err := t.Execute(w, etudiantCo)
		if err != nil {
			log.Printf("error exec template : ", err)
		}
	}
}

func accueil(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de accueil :", r.Method)

	if r.Method == "GET" {
		t, err := template.ParseFiles("./web/html/accueil.html")
		if err != nil {
			fmt.Print("erreur chargement accueil.html")
		}
		t.Execute(w, nil)
	} else if r.Method == "POST" {

		if r.URL.String() == "/login?login" {
			login := r.FormValue("login")
			password := r.FormValue("password")
			fmt.Println("tentative de co avec :", login, " ", password)
			existe := BDD.LoginCorrect(login, password)

			if existe {
				etudiantCo = BDD.GetInfo(login, password)
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
				return
			} else {
				fmt.Println("login incorrecte")
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		} else if r.URL.String() == "/login?register" {
			// pas de vérification de champs implémenter pour l'instant
			etudiantCo = BDD.Etudiant{
				Login:      r.FormValue("login"),
				Password:   r.FormValue("password"),
				Prenom:     r.FormValue("prenom"),
				Nom:        r.FormValue("nom"),
				Mail:       r.FormValue("mail"),
				DefiSucess: 0,
			}
			BDD.Register(etudiantCo)
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
	}
}

func upload() {

}
