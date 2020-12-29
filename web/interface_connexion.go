package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/logs"
	"time"

	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	//"github.com/gomodule/redigo/redis" pas sur de ce truc.
	"html/template"
	"net/http"
)

/**
Fonction pour lancer l'interface web
*/
func InitWeb() {
	http.HandleFunc("/", root)                     // Page de base : http://localhost:8080
	http.HandleFunc("/login", accueil)             // Page d'acceuil : http://localhost:8080/login
	http.HandleFunc("/pageEtudiant", pageEtudiant) // Page étudiant : http://localhost:8080/pageEtudiant
	http.HandleFunc("/pageAdmin", pageAdmin)       // Page admin : http://localhost:8080/pageAdmin
	err := http.ListenAndServe(":8080", nil)       // port utilisé
	if err != nil {
		fmt.Printf("ListenAndServe: ", err)
	}
	setupRoutes()

}

func root(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func accueil(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de accueil :", r.Method)

	if _, err := r.Cookie("token"); err == nil {
		http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		return
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

			fmt.Println("tentative de co avec :", login, " ", password)
			existe := BDD.LoginCorrect(login, password)
			if existe {
				// crée un go routine qui envoie le token, voir si on peut faire ça en même temps que la redirection.
				fmt.Println("Création du token : ")
				token := tokenGenerator()
				expiration := time.Now().Add(1 * time.Minute)
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
				http.SetCookie(w, &cookie)
				fmt.Println("insert login=", login, " token=", token)
				BDD.InsertToken(login, token)

				logs.WriteLog(login, "connexion")
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)

				go DeleteToken(login)
				return
			} else {
				fmt.Println("login incorrecte")
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		} else if r.URL.String() == "/login?register" {
			// request provient du formulaire pour s'enregistrer
			// pas de vérification de champs implémenter pour l'instant
			etu := BDD.Etudiant{
				Login:      r.FormValue("login"),
				Password:   r.FormValue("password"),
				Prenom:     r.FormValue("prenom"),
				Nom:        r.FormValue("nom"),
				Mail:       r.FormValue("mail"),
				DefiSucess: 0,
			}
			BDD.Register(etu) // ajouter l'etudiant dans la base de données.

			logs.WriteLog(etu.Login, "création du compte : "+etu.Login+":"+etu.Password)

			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
	}
}

func DeleteToken(login string) {
	time.Sleep(1 * time.Minute)
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

func setupRoutes() {
	http.HandleFunc("/pageEtudiant", pageEtudiant)
	http.ListenAndServe(":8080", nil)
}

func password() {
	fmt.Print()
}
