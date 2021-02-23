package web

import (
	"crypto/rand"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"time"
)

type dataConnexion struct {
	ConnexionErreur bool
}

func accueil(w http.ResponseWriter, r *http.Request) {
	fmt.Println("methode de accueil :", r.Method)

	if tk, err := r.Cookie("token"); err == nil {
		if BDD.TokenExiste(tk.Value) {
			fmt.Println("Token existe : ", tk.Value)
			role := BDD.TokenRole(tk.Value)
			fmt.Println(role)
			if role == "etudiants" {
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
			} else if role == "administrateur" {
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			}
			return
		}
	}
	if r.Method == "GET" {
		if r.URL.String() == "/login" {
			t, err := template.ParseFiles("./web/html/accueil.html")
			if err != nil {
				fmt.Print("erreur chargement accueil.html")
			}
			_ = t.Execute(w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}

	} else if r.Method == "POST" {

		if r.URL.String() == "/login?login" {
			// request provient du formulaire pour se connecter
			login := r.FormValue("login")
			password := r.FormValue("password")
			fmt.Println("Tentative de connexion avec :", login, " ", password)
			existe := BDD.LoginCorrect(login, password) // on test le couple login/passwordHashé
			if existe {
				//Création du token
				token := tokenGenerator()
				temps := 5 * time.Minute // défini le temps d'attente
				expiration := time.Now().Add(temps)
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
				http.SetCookie(w, &cookie)
				fmt.Println("(login=", login, ",token=", token)
				BDD.InsertToken(login, token)

				logs.WriteLog(login, "connexion étudiant")
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)

				go DeleteToken(login, temps)
				return
			} else {
				logs.WriteLog(login, "mot de passe incorrecte connexion étudiant")
				page, err := template.ParseFiles("./web/html/accueil.html")
				if err != nil {
					fmt.Print("erreur chargement accueil.html")
				} else {
					data := dataConnexion{
						ConnexionErreur: true,
					}
					err = page.Execute(w, data)
				}
			}
		} else if r.URL.String() == "/login?register" {
			// request provient du formulaire pour s'enregistrer
			// pas de vérification de champs implémenter pour l'instant
			if BDD.IsLoginUsed(r.FormValue("login")) {
				fmt.Printf("votre login existe déjà")
				http.Redirect(w, r, "/login", http.StatusFound)
			} else {
				etu := BDD.Etudiant{
					Login:    r.FormValue("login"),
					Password: r.FormValue("password"),
					Prenom:   r.FormValue("prenom"),
					Nom:      r.FormValue("nom"),
					Mail:     r.FormValue("mail"),
				}
				passwordHashed, err := bcrypt.GenerateFromPassword([]byte(etu.Password), 14) // hashage du mot de passe
				if err == nil {
					etu.Password = string(passwordHashed) // le mot de passe à stocker est hashé
					BDD.Register(etu)                     // ajouter l'etudiant dans la base de données.

					logs.WriteLog(etu.Login, "création du compte : "+etu.Login+":"+etu.Password)

					http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
				} else {
					logs.WriteLog(r.FormValue("login"), "échec création du compte")
					http.Redirect(w, r, "/login", http.StatusFound)
				}
			}
		}
	}
}

/*
	Fonction qui permet de gérer la connexion de l'administrateur en fonction de l'url de la requête HTTP client
*/
func connexionAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("methode de connexionAdmin :", r.Method)

	// On vérifie au préalable si sur le client il y a déjà une connexion en cours (token)
	// Si oui on redirige directement vers la page appropriée
	if tk, err := r.Cookie("token"); err == nil {
		if BDD.TokenExiste(tk.Value) {
			fmt.Println("Token existe : ", tk.Value)
			role := BDD.TokenRole(tk.Value)
			fmt.Println(role)
			if role == "etudiants" {
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
			} else if role == "administrateur" {
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			}
			return
		}
	}

	if r.Method == "GET" {

		// Lorsqu'il s'agit d'une requête avec la méthode GET on parse l'url
		if r.URL.String() == "/loginAdmin" {
			page, err := template.ParseFiles("./web/html/connexionAdmin.html")
			if err != nil {
				fmt.Print("erreur chargement connexionAdmin.html : ", err)
			} else {
				data := dataConnexion{
					ConnexionErreur: false,
				}
				err = page.Execute(w, data)
				if err != nil {
					fmt.Print("erreur affichage connexionAdmin.html : ", err)
				}
			}
		} else {
			http.Redirect(w, r, "/loginAdmin", http.StatusFound)
		}

	} else if r.Method == "POST" {

		if r.URL.String() == "/loginAdmin?login" {

			login := r.FormValue("login")
			password := r.FormValue("password")
			fmt.Println("Tentative de connexion admin avec :", login, " ", password)
			existe := BDD.LoginCorrectAdmin(login, password) // on test le couple login/passwordHashé
			if existe {
				//Création du token
				token := tokenGenerator()
				temps := 5 * time.Minute // défini le temps d'attente
				expiration := time.Now().Add(temps)
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
				http.SetCookie(w, &cookie)
				BDD.InsertToken(login, token)
				logs.WriteLog(login, "connexion admin réussie, création d'un token")
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
				go DeleteToken(login, temps)
				return
			} else {
				logs.WriteLog(login, "mot de passe incorrecte connexion admin")
				page, err := template.ParseFiles("./web/html/connexionAdmin.html")
				if err != nil {
					fmt.Print("erreur chargement connexionAdmin.html")
				} else {
					data := dataConnexion{
						ConnexionErreur: true,
					}
					err = page.Execute(w, data)
					if err != nil {
						fmt.Println("erreur affichage connexionAdmin.html : ", err)
					}
				}
			}
		}
	}
}

/*
	Fonction qui permet de supprimer de la BD le token d'une connexion après une durée donnée
*/
func DeleteToken(login string, temps time.Duration) {
	time.Sleep(temps)
	logs.WriteLog(login, "déconnexion du serveur")
	BDD.DeleteToken(login)
	return
}

/*
	Fonction qui permet de générer un token
	@return une chaîne de caractères de la forme : %x[.., .., .., ..]
*/
func tokenGenerator() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprint("%x", b)
}
