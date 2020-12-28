package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

type data_pageEtudiant struct {
	UserInfo BDD.Etudiant
	Res      string
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
	data := data_pageEtudiant{
		UserInfo: etu,
		Res:      "",
	}
	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {
		//Charge la template html
		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))

		// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
		if err := t.Execute(w, data); err != nil {
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
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
	}
}
