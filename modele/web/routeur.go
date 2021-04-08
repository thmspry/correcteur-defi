package web

import (
	"encoding/json"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"net/http"
)

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
	http.HandleFunc("/GetParticipantsDefis", GetParticipantsDefis)
	http.HandleFunc("/", Redirection) // Redirection url par défaut : http://localhost:8192/

	err := http.ListenAndServe(":8192", nil) // port utilisé
	if err != nil {
		fmt.Printf("ListenAndServe: ", err)
	}
}

func Redirection(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func GetDefis(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie("token"); err != nil || !DAO.TokenExiste(token.Value) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(DAO.GetDefis())
}

func GetDefiActuel(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie("token"); err != nil || !DAO.TokenExiste(token.Value) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(DAO.GetDefiActuel())
}

func GetParticipantsDefis(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie("token"); err != nil || !DAO.TokenExiste(token.Value) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	etudiantsNb := len(DAO.GetEtudiants())
	nbDefiActuel := DAO.GetDefiActuel().Num
	defis := DAO.GetDefis()
	participants := make([]modele.StatsDefi, 0)
	for _, defi := range defis {
		nbReussi := len(DAO.GetResultatsByEtat(defi.Num, 1))
		moyenne := 0

		if nbReussi != 0 {
			for _, result := range DAO.GetResultatsByEtat(defi.Num, 1) {
				moyenne += result.Tentative
			}
			moyenne = moyenne / len(DAO.GetResultatsByEtat(defi.Num, 1))
		}
		participants = append(participants, modele.StatsDefi{
			Num:               defi.Num,
			ParticipantsDefi:  len(DAO.GetParticipants(defi.Num)),
			Reussite:          nbReussi,
			MoyenneTentatives: moyenne,
		})
	}
	json.NewEncoder(w).Encode(modele.StatsDefis{NbEtudiants: etudiantsNb, NbDefiActuel: nbDefiActuel, Participants: participants})
}
