package web

import (
	"encoding/json"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"net/http"
)

var domain = "go-correcteur.com"

/**
Fonction pour lancer l'interface web
*/
func InitWeb() {

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web")))) // compliqué à expliquer
	http.HandleFunc("/login", accueil)                                                // Page d'acceuil : http://localhost:8192/login
	http.HandleFunc("/loginAdmin", connexionAdmin)                                    // Page de connexion admin : http://localhost:8192/loginAdmin
	http.HandleFunc("/pageEtudiant", pageEtudiant)                                    // Page étudiant : http://localhost:8192/pageEtudiant
	http.HandleFunc("/pageAdmin", pageAdmin)                                          // Page admin : http://localhost:8192/pageAdmin
	http.HandleFunc("/GetDefis", GetDefis)
	http.HandleFunc("/GetDefiActuel", GetDefiActuel)
	http.HandleFunc("/", Redirection)

	err := http.ListenAndServe(":8192", nil) // port utilisé
	if err != nil {
		fmt.Printf("ListenAndServe: ", err)
	}
}

func Redirection(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func GetDefis(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie("token"); err != nil || !BDD.TokenExiste(token.Value) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(BDD.GetDefis())
}

func GetDefiActuel(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie("token"); err != nil || !BDD.TokenExiste(token.Value) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(BDD.GetDefiActuel())
}
