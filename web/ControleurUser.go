package web

import (
	"bufio"
	"fmt"
	"github.com/aodin/date"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/testeur"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type data_pageEtudiant struct { // Données transmises à la page Etudiant
	UserInfo      config.Etudiant
	Defi_sent     bool
	Defi_actuel   config.Defi
	Resultat_defi config.ResBDD
	ResTest       []config.Resultat
	Msg_res       string
	Script        []string
}

/**
Fonction pour afficher la page Etudiant à l'adresse /pageEtudiant
*/
func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("methode de pageEtudiant :", r.Method)
	num_defi_actuel := BDD.GetDefiActuel().Num

	//Si il y a n'y a pas de token dans les cookies alors l'utilisateur est redirigé vers la page de login
	if token, err := r.Cookie("token"); err != nil || !BDD.TokenExiste(token.Value) {
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
		ResTest:     BDD.GetResultatActuel(etu.Login),
	}

	if data.Defi_actuel.Num != -1 {
		if !date.Today().Within(date.NewRange(data.Defi_actuel.Date_debut, data.Defi_actuel.Date_fin)) {
			data.Defi_actuel.Num = -1
		}
	}

	if data.Defi_actuel.Num != -1 {
		data.Defi_sent = manipStockage.Contains(config.Path_scripts, "script_"+etu.Login+"_"+strconv.Itoa(data.Defi_actuel.Num))
		if data.Defi_sent {
			data.Resultat_defi = BDD.GetResult(etu.Login, data.Defi_actuel.Num)
		}
	}

	f, err := os.Open(config.Path_scripts + "script_" + etu.Login + "_" + strconv.Itoa(data.Defi_actuel.Num))
	if err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			data.Script = append(data.Script, scanner.Text())
		}
	}
	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {
		//Charge la template html

		if r.URL.Query()["logout"] != nil {
			DeleteToken(etu.Login, time.Second*0)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.URL.Query()["test"] != nil {
			//data.Msg_res, data.ResTest = testeur.Test(etu.Login)
			data.Msg_res, data.ResTest = testeur.TestArtificiel("E197051L")
		}

		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))

		// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
		if err := t.Execute(w, data); err != nil {
			log.Printf("error exec template : ", err.Error())
		}

		//Si la méthode est post c'est qu'on vient d'envoyer un fichier pour le faire tester
	} else if r.Method == "POST" {
		if r.URL.Query()["upload"] != nil {

			r.ParseMultipartForm(10 << 20)

			file, _, _ := r.FormFile("script_etu")

			script, _ := os.Create(config.Path_scripts + "script_" + etu.Login + "_" + strconv.Itoa(num_defi_actuel)) // remplacer handler.Filename par le nom et on le place où on veut
			BDD.SaveResultat(etu.Login, num_defi_actuel, -1, nil, false)

			_, err = io.Copy(script, file)

			os.Chmod(config.Path_scripts+"script_"+etu.Login+"_"+strconv.Itoa(num_defi_actuel), 770)
			file.Close()
			script.Close()
			logs.WriteLog(etu.Login, "upload de script du défis "+strconv.Itoa(num_defi_actuel))
			data.Msg_res, data.ResTest = testeur.Test(etu.Login)
			t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))
			// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
			if err := t.Execute(w, data); err != nil {
				log.Printf("error exec template : ", err.Error())
			}
		}
	}
}
