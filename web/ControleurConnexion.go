package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"time"

	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	//"github.com/gomodule/redigo/redis" pas sur de ce truc.
	"html/template"
	"net/http"
)

func accueil(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de accueil :", r.Method)

	if tk, err := r.Cookie("token"); err == nil {
		if BDD.TokenExiste(tk.Value) {
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
			return
		}
	}
	if r.Method == "GET" {

		t, err := template.ParseFiles("./web/html/accueil.html")
		if err != nil {
			fmt.Print("erreur chargement accueil.html")
		}
		_ = t.Execute(w, nil)
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

				logs.WriteLog(login, "connexion")
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)

				go DeleteToken(login, temps)
				return
			} else {
				fmt.Println("login '" + login + "' incorrecte")
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		} else if r.URL.String() == "/login?register" {
			// request provient du formulaire pour s'enregistrer
			// pas de vérification de champs implémenter pour l'instant
			etu := BDD.Etudiant{
				Login:    r.FormValue("login"),
				Password: r.FormValue("password"),
				Prenom:   r.FormValue("prenom"),
				Nom:      r.FormValue("nom"),
				Mail:     r.FormValue("mail"),
			}
			passwordHashed, err := bcrypt.GenerateFromPassword([]byte(etu.Password), 14) // hashage du mot de passe
			if err == nil {
				etu.Password = string(passwordHashed) // le mot de passe à stocké est hashé
				BDD.Register(etu)                     // ajouter l'etudiant dans la base de données.

				logs.WriteLog(etu.Login, "création du compte : "+etu.Login+":"+etu.Password)

				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
			} else {
				// renvoi vers un page error/signification d'une erreur
			}

		}
	}
}

func DeleteToken(login string, temps time.Duration) {
	time.Sleep(temps)
	logs.WriteLog(login, "déconnexion du serveur")
	BDD.DeleteToken(login)
	return
}

//On genere un token.
func tokenGenerator() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprint("%x", b)
}

//Pour plus tard, essayer de hasher le motdepasse pour ne pas le stocker en clair.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
