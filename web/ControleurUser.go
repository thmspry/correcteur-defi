package web

import (
	"bufio"
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
	Error        bool
	ErrorMsg     string
	NbTestReussi int
	NbTestEchoue int
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
		Error:        false,
		ErrorMsg:     "",
		NbTestReussi: 0,
		NbTestEchoue: 0,
	}

	if data.DefiActuel.Num != 0 {
		data.DefiSent = manipStockage.Contains(modele.PathScripts, "script_"+etu.Login+"_"+strconv.Itoa(data.DefiActuel.Num))
	}

	f, err := os.Open(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(data.DefiActuel.Num))
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
			//data.MsgRes, data.ResTest = testeur.Test(etu.Login)
			//data.MsgRes, data.ResTest = testeur.TestArtificielReussite("E197051L")
			data.MsgRes, data.ResTest = testeur.TestArtificielEchec("E197051L")
			for _, test := range data.ResTest {
				if test.Etat == 1 {
					data.NbTestReussi++
				}
			}
			data.NbTestEchoue = len(data.ResTest) - data.NbTestReussi

			//DAO.SaveResultat("E197051L", 1, 1, data.ResTest, false)
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
			b, _ := ioutil.ReadFile(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(numDefiActuel))
			contenuscript := string(b)
			fmt.Printf(contenuscript)
			value := strings.Contains(contenuscript, "!bin/bash")
			if value == true {
				fmt.Printf("ok")
			} else {
				fmt.Printf("ko")
			}

			script, _ := os.Create(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(numDefiActuel)) // remplacer handler.Filename par le nom et on le place où on veut
			DAO.SaveResultat(etu.Login, numDefiActuel, -1, nil, false)

			_, err = io.Copy(script, file) //on l'enregistre dans notre système de fichier
			fmt.Println(modele.PathScripts + "script_" + etu.Login + "_" + strconv.Itoa(numDefiActuel))

			os.Chmod(modele.PathScripts+"script_"+etu.Login+"_"+strconv.Itoa(numDefiActuel), 770) //change le chmode du fichier
			file.Close()
			script.Close()
			logs.WriteLog(etu.Login, "upload de script du défis "+strconv.Itoa(numDefiActuel))
			data.MsgRes, data.ResTest = testeur.Test(etu.Login)
			t := template.Must(template.ParseFiles("./web/html/pageEtudiant.html"))
			// execute la page avec la structure "etu" qui viendra remplacer les éléments de la page en fonction de l'étudiant (voir pageEtudiant.html)
			if err := t.Execute(w, data); err != nil {
				log.Printf("error exec template : ", err.Error())
			}
		}
	}
}
