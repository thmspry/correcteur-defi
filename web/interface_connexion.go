package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"io/ioutil"
	"time"

	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	//"github.com/gomodule/redigo/redis" pas sur de ce truc.
	"html/template"
	"log"
	"net/http"
)

var etudiantCo BDD.Etudiant

func InitWeb() {

	http.HandleFunc("/login", accueil)             // Page d'acceuil : http://localhost:8080/login
	http.HandleFunc("/pageEtudiant", pageEtudiant) // Page étudiant : http://localhost:8080/pageEtudiant

	err := http.ListenAndServe(":8080", nil) // port utilisé
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	setupRoutes()
}

func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de pageEtudiant :", r.Method)
	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {

		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))
		fmt.Println(r.Cookie("token"))

		token, _ := r.Cookie("token")
		name := BDD.GetNameByToken(token.Value)
		fmt.Println("name récup de getnamebyToken", name)
		etu := BDD.GetInfo(name)
		err := t.Execute(w, etu)
		if err != nil {
			log.Printf("error exec template : ", err)
		}
	} else if r.Method == "POST" {
		fmt.Printf("pageEtudiant post")
		if r.URL.String() == "/pageEtudiant" {
			// TODO récupération du fichier (qui se trouve dans r.FormValue["uploadfile"]
			fmt.Println("File Upload Endpoint Hit")

			// Parse our multipart form, 10 << 20 specifies a maximum
			// upload of 10 MB files.
			r.ParseMultipartForm(10 << 20)
			// FormFile returns the first file for the given key `myFile`
			// it also returns the FileHeader so we can get the Filename,
			// the Header and the size of the file
			file, handler, err := r.FormFile("myFile")
			if err != nil {
				fmt.Println("Error Retrieving the File")
				fmt.Println(err)
				return
			}
			defer file.Close()
			fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			fmt.Printf("File Size: %+v\n", handler.Size)
			fmt.Printf("MIME Header: %+v\n", handler.Header)

			// Create a temporary file within our directory that follows
			// a particular naming pattern
			tempFile, err := ioutil.TempFile("ressource/script_etudiants", "oh.txt")
			if err != nil {
				fmt.Println(err)
			}
			defer tempFile.Close()

			// read all of the contents of our uploaded file into a
			// byte array
			fileBytes, err := ioutil.ReadAll(file)
			if err != nil {
				fmt.Println(err)
			}
			// write this byte array to our temporary file
			tempFile.Write(fileBytes)
			// return that we have successfully uploaded our file!
			fmt.Fprintf(w, "Successfully Uploaded File\n")
		}
	}
}

func accueil(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de accueil :", r.Method)

	if r.Method == "GET" {
		if token, _ := r.Cookie("token"); token != nil {
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
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
				BDD.InsertToken(login, token)

				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)

				go DeleteToken(token)
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

func DeleteToken(token string) {
	time.Sleep(1 * time.Minute)
	BDD.DeleteToken(token)
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
