package web

import (
	"bufio"
	"fmt"
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
	numDefiActuel := BDD.GetDefiActuel().Num

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
		UserInfo:      etu,
		Defi_actuel:   BDD.GetDefiActuel(),
		ResTest:       BDD.GetResultatActuel(etu.Login),
		Resultat_defi: BDD.GetResult(etu.Login, numDefiActuel),
	}
	if data.Defi_actuel.Num != -1 {
		if time.Now().Sub(data.Defi_actuel.DateDebut) < 0 || time.Now().Sub(data.Defi_actuel.DateFin) > 0 {
			data.Defi_actuel.Num = -1
		}
	}
	if data.Defi_actuel.Num != -1 {
		data.Defi_sent = manipStockage.Contains(config.PathScripts, "script_"+etu.Login+"_"+strconv.Itoa(data.Defi_actuel.Num))
	}

	f, err := os.Open(config.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(data.Defi_actuel.Num))
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
			data.Msg_res, data.ResTest = testeur.Test(etu.Login)
			//data.Msg_res, data.ResTest = testeur.TestArtificiel("E197051L")
		}

		t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))

		// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
		if err := t.Execute(w, data); err != nil {
			log.Printf("error exec template : ", err.Error())
			logs.WriteLog("Erreur d'execution pageEtudiant.html : ", err.Error())
		}

		//Si la méthode est post c'est qu'on vient d'envoyer un fichier pour le faire tester
	} else if r.Method == "POST" {
		if r.URL.Query()["upload"] != nil {

			r.ParseMultipartForm(10 << 20) //sert à télécharger des fichiers et le stock sur le serveur

			file, _, _ := r.FormFile("script_etu") // sert à obtenir le descripteur de fichier

			script, _ := os.Create(config.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(numDefiActuel)) // remplacer handler.Filename par le nom et on le place où on veut
			BDD.SaveResultat(etu.Login, numDefiActuel, -1, nil, false)

			_, err = io.Copy(script, file) //on l'enregistre dans notre système de fichier
			fmt.Println("teststetesttestestestesteste")
			//b, _ := ioutil.ReadFile(script.Name())
			//fmt.Printf(string(b))
			//fmt.Printf(ioutil.ReadFile(script.Name()))
			os.Chmod(config.PathScripts+"script_"+etu.Login+"_"+strconv.Itoa(numDefiActuel), 770) //change le chmode du fichier
			file.Close()
			script.Close()
			logs.WriteLog(etu.Login, "upload de script du défis "+strconv.Itoa(numDefiActuel))
			data.Msg_res, data.ResTest = testeur.Test(etu.Login)
			t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))
			// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
			if err := t.Execute(w, data); err != nil {
				log.Printf("error exec template : ", err.Error())
			}
		}
	}
}
