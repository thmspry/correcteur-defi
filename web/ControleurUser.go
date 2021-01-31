package web

import (
	"fmt"
	"github.com/aodin/date"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type data_pageEtudiant struct {
	UserInfo      BDD.Etudiant
	Defi_sent     bool
	Defi_actuel   BDD.Defi
	Resultat_defi BDD.ResBDD
	ResTest       []testeur.Resultat
	Msg_res       string
}

/**
Fonction pour afficher la page Etudiant à l'adresse /pageEtudiant
*/
func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method de pageEtudiant :", r.Method)
	num_defi_actuel := BDD.GetDefiActuel().Num

	//Si il y a n'y a pas de token dans les cookies alors l'utilisateur est redirigé vers la page de login
	if _, err := r.Cookie("token"); err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	token, _ := r.Cookie("token")            //récupère le token du cookie
	login := BDD.GetNameByToken(token.Value) // récupère le login correspondant au token
	etu := BDD.GetEtudiant(login)            // récupère les informations de l'étudiant grâce au login

	//Parse data
	data := data_pageEtudiant{
		UserInfo:    etu,
		Defi_actuel: BDD.GetDefiActuel(),
		ResTest:     nil,
	}

	if data.Defi_actuel.Num != -1 {
		if !date.Today().Within(date.NewRange(data.Defi_actuel.Date_debut, data.Defi_actuel.Date_fin)) {
			data.Defi_actuel.Num = -1
		}
	}

	if data.Defi_actuel.Num == -1 {
		data.Defi_sent = testeur.Contains(config.Path_scripts, "script_"+etu.Login+"_"+strconv.Itoa(data.Defi_actuel.Num)+".sh")
		if data.Defi_sent {
			data.Resultat_defi = BDD.GetResult(etu.Login, data.Defi_actuel.Num)
		}
	}

	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {
		//Charge la template html

		if r.URL.Query()["logout"] != nil {
			fmt.Println("logout " + etu.Login)
			DeleteToken(etu.Login, time.Second*0)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.URL.Query()["test"] != nil {
			data.Msg_res, data.ResTest = testeur.Test(etu.Login)
		}

		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))

		// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
		if err := t.Execute(w, data); err != nil {
			log.Printf("error exec template : ", err)
		}

		//Si la méthode est post c'est qu'on vient d'envoyer un fichier pour le faire tester
	} else if r.Method == "POST" {
		fmt.Printf("pageEtudiant post")
		if r.URL.Query()["upload"] != nil {

			r.ParseMultipartForm(10 << 20)

			file, _, err := r.FormFile("script_etu")
			if err != nil {
				fmt.Println("Error Retrieving the File")
				fmt.Println(err)
				return
			}
			defer file.Close()

			script, err := os.Create(config.Path_scripts + "script_" + etu.Login + "_" + strconv.Itoa(num_defi_actuel) + ".sh") // remplacer handler.Filename par le nom et on le place où on veut

			BDD.SaveResultat(etu.Login, num_defi_actuel, -1, false)
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

			os.Chmod(config.Path_scripts+"script_"+etu.Login+"_"+strconv.Itoa(num_defi_actuel)+".sh", 770)

			logs.WriteLog(etu.Login, "upload de script du défis "+strconv.Itoa(num_defi_actuel))
			data.Msg_res, data.ResTest = testeur.Test(etu.Login)
			http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
		}
	}
}
