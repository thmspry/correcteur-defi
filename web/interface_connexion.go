package web

import (
	"bytes"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
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
	} else if r.Method == "POST" {
		fmt.Printf("sdxlkhbfsdlkjgh")
		if r.URL.String() == "/pageEtudiant?uploadFile" {
			// TODO récupération du fichier (qui se trouve dans r.FormValue["uploadfile"]
			fmt.Println(r.FormValue("uploadfile"))
			_ = r.ParseMultipartForm(32 << 20)
			file, handler, err := r.FormFile("uploadfile")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer file.Close()
			_, _ = fmt.Fprintf(w, "%v", handler.Header)
			f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			_, _ = io.Copy(f, file)
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
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
		_ = t.Execute(w, nil)
	} else if r.Method == "POST" {

		if r.URL.String() == "/login?login" {
			// request provient du formulaire pour se connecter
			login := r.FormValue("login")
			password := r.FormValue("password")
			fmt.Println("tentative de co avec :", login, " ", password)
			existe := BDD.LoginCorrect(login, password)

			if existe {
				etudiantCo = BDD.GetInfo(login)
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
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
			BDD.Register(etudiantCo)
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
	}
}

func upload(filename string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	_ = bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}

// sample usage
func main() {
	target_url := "http://localhost:8080/pageEtudiant?uploadFile"
	filename := "./a.txt"
	_ = upload(filename, target_url)
}
