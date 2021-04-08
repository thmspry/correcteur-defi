package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/testeur"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type data_pageEtudiant struct { // Données transmises à la page Etudiant
	UserInfo     modele.Etudiant
	DefiSent     bool
	DefiActuel   modele.Defi
	ResultatDefi modele.Resultat
	ResTest      []modele.ResultatTest
	MsgRes       string
	Script       []string
	Alert        string
	NbTestReussi int
	NbTestEchoue int
	Classement   []modele.Resultat
}

/**
Fonction pour afficher la page Etudiant à l'adresse /pageEtudiant
*/
func pageEtudiant(w http.ResponseWriter, r *http.Request) {
	fmt.Println("methode de pageEtudiant :", r.Method)
	numDefiActuel := DAO.GetDefiActuel().Num

	//Si il y a n'y a pas de token dans les cookies alors l'utilisateur est redirigé vers la page de login
	if token, err := r.Cookie("token"); err != nil || !DAO.TokenExiste(token.Value) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	token, _ := r.Cookie("token")             //récupère le token du cookie
	login := DAO.GetLoginByToken(token.Value) // récupère le login correspondant au token
	etu := DAO.GetEtudiant(login)             // récupère les informations de l'étudiant grâce au login

	//Parse data
	data := data_pageEtudiant{
		UserInfo:     etu,
		DefiActuel:   DAO.GetDefiActuel(),
		ResTest:      DAO.GetResultatTest(etu.Login),
		ResultatDefi: DAO.GetResult(etu.Login, numDefiActuel),
		NbTestReussi: 0,
		NbTestEchoue: 0,
		Classement:   DAO.GetClassement(numDefiActuel),
	}

	if data.DefiActuel.Num != 0 {
		data.DefiSent = manipStockage.Contains(modele.PathScripts, "script_"+etu.Login+"_"+strconv.Itoa(data.DefiActuel.Num))
	}

	data.Script = manipStockage.GetFileLineByLine(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(data.DefiActuel.Num))
	//Check la méthode utilisé par le formulaire
	if r.Method == "GET" {
		//Charge la template html
		if r.URL.Query()["logout"] != nil {
			DeleteToken(etu.Login, time.Second*0)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.URL.Query()["test"] != nil {
			data.MsgRes, data.ResTest = testeur.Test(etu.Login)
			for _, test := range data.ResTest {
				if test.Etat == 1 {
					data.NbTestReussi++
				}
			}
			data.NbTestEchoue = len(data.ResTest) - data.NbTestReussi
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
			path := modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(numDefiActuel)
			script, _ := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut

			io.Copy(script, file) //on l'enregistre dans notre système de fichier
			//os.Chmod(modele.PathScripts+"script_"+etu.Login+"_"+strconv.Itoa(numDefiActuel), 770) //change le chmode du fichier (marche pas sous windows)
			file.Close()
			script.Close()
			b, _ := ioutil.ReadFile(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(numDefiActuel))
			contenuscript := string(b)
			if strings.Contains(contenuscript, "#!/bin/bash") {
				data.Alert = "upload du script pour le défi " + strconv.Itoa(numDefiActuel)
				logs.WriteLog(etu.Login, data.Alert)
				data.MsgRes, data.ResTest = testeur.Test(etu.Login)
				DAO.SaveResultat(etu.Login, numDefiActuel, -1, nil, false)
			} else {
				os.Remove(path)
				data.Alert = "le script n'a pas été upload car il ne contient pas '!/bin/bash'"
			}
			data.Script = manipStockage.GetFileLineByLine(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(data.DefiActuel.Num))
			data.MsgRes, data.ResTest = testeur.Test(data.UserInfo.Login)

			t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))
			// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
			if err := t.Execute(w, data); err != nil {
				log.Printf("error exec template : ", err.Error())
			}
		}
	}
}
