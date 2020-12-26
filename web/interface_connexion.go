package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	//"github.com/gomodule/redigo/redis" pas sur de ce truc.
	"html/template"
	"log"
	"net/http"
)

var etudiantCo BDD.Etudiant

/**
Fonction pour lancer l'interface web
*/
func InitWeb() {
	http.HandleFunc("/", accueil)                  // Page de base : http://localhost:8080
	http.HandleFunc("/login", accueil)             // Page d'acceuil : http://localhost:8080/login
	http.HandleFunc("/pageEtudiant", pageEtudiant) // Page étudiant : http://localhost:8080/pageEtudiant
	http.HandleFunc("/pageAdmin", pageAdmin)       // Page admin : http://localhost:8080/pageAdmin
	err := http.ListenAndServe(":8080", nil)       // port utilisé
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	setupRoutes()

}

/**
Fonction pour afficher la page Etudiant à l'adresse /pageEtudiant
*/
func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de pageEtudiant :", r.Method)

	//Si il y a n'y a pas de token dans les cookies alors l'utilisateur est redirigé vers la page de login
	if _, err := r.Cookie("token"); err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	token, _ := r.Cookie("token")            //récupère le token du cookie
	login := BDD.GetNameByToken(token.Value) // récupère le login correspondant au token
	etu := BDD.GetInfo(login)                // récupère les informations de l'étudiant grâce au login

	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {
		//Charge la template html
		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))
		// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
		if err := t.Execute(w, etu); err != nil {
			log.Printf("error exec template : ", err)
		}

		//Si la méthode est post c'est qu'on vient d'envoyer un fichier pour le faire tester
	} else if r.Method == "POST" {
		fmt.Printf("pageEtudiant post")
		if r.URL.String() == "/pageEtudiant" {

			r.ParseMultipartForm(10 << 20)

			file, _, err := r.FormFile("script_etu")
			if err != nil {
				fmt.Println("Error Retrieving the File")
				fmt.Println(err)
				return
			}
			defer file.Close()

			num, _ := testeur.Defi_actuel()
			script, err := os.Create("./ressource/script_etudiants/script_" + etu.Login + "_" + num + ".sh") // remplacer handler.Filename par le nom et on le place où on veut

			if err != nil {
				fmt.Println("Internal Error")
				fmt.Println(err)
				return
			}

			defer script.Close()

			_, err = io.Copy(script, file)
			if err != nil {
				fmt.Println("Internal Error")
				fmt.Println(err)
				return
			}

			// return that we have successfully uploaded our file!
			fmt.Println("Successfully Uploaded File\n")
			//rename fonctionne pas jsp pk
			//os.Rename(handler.Filename, "script_E1000.sh")
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
	}
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
			//existe := true
			if existe {
				// crée un go routine qui envoie le token, voir si on peut faire ça en même temps que la redirection.
				fmt.Println("Création du token : ")
				token := tokenGenerator()
				expiration := time.Now().Add(1 * time.Minute)
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
				http.SetCookie(w, &cookie)
				fmt.Println("insert login=", login, " token=", token)
				BDD.InsertToken(login, token)

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
			etudiantCo = BDD.Etudiant{
				Login:      r.FormValue("login"),
				Password:   r.FormValue("password"),
				Prenom:     r.FormValue("prenom"),
				Nom:        r.FormValue("nom"),
				Mail:       r.FormValue("mail"),
				DefiSucess: 0,
			}
		}
		BDD.Register(etudiantCo) // ajouter l'etudiant dans la base de données.
		http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
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
