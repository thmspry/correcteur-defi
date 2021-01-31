package web

import (
	"bufio"
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
)

type data_pageAdmin struct {
	Etu_select  string
	Etudiants   []BDD.Etudiant
	Res_etu     []BDD.ResBDD
	ListeDefis  []BDD.Defi
	File        []string
	Defi_actuel BDD.Defi
	Logs        []string
	Log         []string
	LogDate     string
}

func pageAdmin(w http.ResponseWriter, r *http.Request) {
	data := data_pageAdmin{
		Etudiants:   BDD.GetEtudiants(),
		Defi_actuel: BDD.GetDefiActuel(),
		ListeDefis:  BDD.GetDefis(),
		Logs:        testeur.GetFiles(config.Path_log),
	}
	data.Logs = data.Logs[:len(data.Logs)-1]

	//if date actuelle > defi actel.datefin alors defiactuel.num = -1
	if data.Defi_actuel.Num != -1 {
		if !date.Today().Within(date.NewRange(data.Defi_actuel.Date_debut, data.Defi_actuel.Date_fin)) {
			data.Defi_actuel.Num = -1
		}
	}
	fmt.Println(r.URL.String())
	if r.Method == "GET" {

		//Permet d'afficher les logs d'une date précise
		if r.URL.Query()["Log"] != nil {
			log := r.URL.Query()["Log"][0]
			data.LogDate = log
			f, err := os.Open(config.Path_log + log)
			if err != nil {
				data.File[0] = "erreur pour récupérer le script de l'étudiant"
			} else {
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					data.Log = append(data.Log, scanner.Text())
				}
			}
		}

		//Permet d'afficher les résultats correspondant à un étudiant en particulier
		if r.URL.Query()["Etudiant"] != nil {
			etu := r.URL.Query()["Etudiant"][0]
			data.Etu_select = etu

			//Permet de changer l'état de la du défis
			if r.URL.Query()["Script"] != nil && r.URL.Query()["Etat"] != nil {
				etat := r.URL.Query()["Etat"][0]
				num, _ := strconv.Atoi(r.URL.Query()["Script"][0])
				if etat == "1" {
					BDD.SaveResultat(etu, num, 0, true)
				} else {
					BDD.SaveResultat(etu, num, 1, true)
				}

				//Permet d'afficher le contenu du script envoyé par l'étudiant pour le défi séléctionné
			} else if r.URL.Query()["Script"] != nil {
				num := r.URL.Query()["Script"][0]

				f, err := os.Open(config.Path_scripts + "script_" + etu + "_" + num + ".sh")
				if err != nil {
					data.File[0] = "erreur pour récupérer le script de l'étudiant"
				} else {
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						data.File = append(data.File, scanner.Text())
					}
				}
			}
			data.Res_etu = BDD.GetAllResultat(etu)
		}

		t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))

		if err := t.Execute(w, data); err != nil {
			log.Printf("error exec template : ", err)
		}
	}

	if r.Method == "POST" {

		if r.URL.Query()["form"][0] == "modify_date" {
			logs.WriteLog("Admin", "modification de la date de rendu")
			debut, err1 := date.Parse(r.FormValue("date_debut"))
			fin, err2 := date.Parse(r.FormValue("date_fin"))
			if err1 != nil || err2 != nil {
				fmt.Println("Erreur de format dans les dates entrés pour modifier la date")
			} else {
				BDD.ModifyDefi(BDD.GetDefiActuel().Num, debut, fin)
			}
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}
		//Permet de récupérer les résultats de tous les étudiants ainsi que leurs informations pour un défi donné
		if r.URL.Query()["form"][0] == "getResult" {
			num := r.FormValue("num")
			n, err := strconv.Atoi(num)
			if err != nil {
				fmt.Println(err.Error())
			}
			file_name := "resultat_" + num + ".csv"
			testeur.CreateCSV(file_name, n)
			w.Header().Set("Content-Disposition", "attachment; filename="+file_name)
			w.Header().Set("Content-Type", "application/octet-stream")
			http.ServeFile(w, r, file_name)
			os.Remove(file_name)
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}

		r.ParseMultipartForm(10 << 20)
		file, _, err := r.FormFile("defi")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()

		defi_actuel := BDD.GetDefiActuel()
		num_defi_actuel := defi_actuel.Num
		path := ""

		if r.URL.Query()["form"][0] == "defi" {
			submit := r.FormValue("submit")
			date_debut, err := date.Parse(r.FormValue("date_debut"))
			if err != nil {
				fmt.Println("Erreur dans le format de la date de début")
			}
			date_fin, err := date.Parse(r.FormValue("date_fin"))
			if err != nil {
				fmt.Println("Erreur dans le format de la date de fin")
			}
			if submit == "modifier" {
				logs.WriteLog("Admin", "modification de la correction")
				BDD.ModifyDefi(BDD.GetDefiActuel().Num, date_debut, date_fin)
				path = config.Path_defis + "correction_" + strconv.Itoa(num_defi_actuel) + ".sh"
			} else {
				logs.WriteLog("Admin", "ajout d'un nouveau défis")
				// ajouter a la table défis
				BDD.AddDefi(num_defi_actuel+1, date_debut, date_fin)
				os.Mkdir(config.Path_jeu_de_tests+"test_defi_"+strconv.Itoa(num_defi_actuel+1), os.ModePerm)
				num_defi_actuel = num_defi_actuel + 1
				path = config.Path_defis + "correction_" + strconv.Itoa(num_defi_actuel) + ".sh"
			}
		} else if r.URL.Query()["form"][0] == "test" {

			logs.WriteLog("Admin", "upload d'un test pour le défi n°"+strconv.Itoa(num_defi_actuel))
			path = config.Path_jeu_de_tests + "test_defi_" + strconv.Itoa(num_defi_actuel) + "/"
			num_test := testeur.Nb_test(path)
			path = config.Path_jeu_de_tests + "test_defi_" + strconv.Itoa(num_defi_actuel) + "/test_" + strconv.Itoa(num_test)
		}

		script, err := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut

		if err != nil {
			fmt.Println("Internal Error")
			fmt.Println(err)
		}

		defer script.Close()

		_, err = io.Copy(script, file)
		if err != nil {
			fmt.Println("Internal Error")
			fmt.Println(err)
		}

		os.Chmod(path, 770)

		// return that we have successfully uploaded our file!
		fmt.Println("Successfully Uploaded File\n")
		//rename fonctionne pas jsp pk
		//os.Rename(handler.Filename, "script_E1000.sh")
		http.Redirect(w, r, "/pageAdmin", http.StatusFound)
	}

}
