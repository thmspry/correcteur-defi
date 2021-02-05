package web

import (
	"fmt"
	"net/http"
)

/**
Fonction pour lancer l'interface web
*/
func InitWeb() {

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web")))) // compliqué à expliquer
	http.HandleFunc("/login", accueil)                                                // Page d'acceuil : http://localhost:8192/login
	http.HandleFunc("/pageEtudiant", pageEtudiant)                                    // Page étudiant : http://localhost:8192/pageEtudiant
	http.HandleFunc("/pageAdmin", pageAdmin)                                          // Page admin : http://localhost:8192/pageAdmin
	err := http.ListenAndServe(":8192", nil)                                          // port utilisé
	if err != nil {
		fmt.Printf("ListenAndServe: ", err)
	}
}
