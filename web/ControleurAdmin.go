package web

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aodin/date"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type data_pageAdmin struct {
	EtuSelect     string
	DefiNumSelect string
	Etudiants     []BDD.Etudiant
	Res_etu       []BDD.ResBDD
	ListeDefis    []BDD.Defi
	File          []string
	Defi_actuel   BDD.Defi
	Participants  []BDD.ParticipantDefi

	Logs    []string
	Log     []string
	LogDate string
}

type SenderData struct {
	FromMail string `json:"fromMail"`
	Username string `json:"username"`
	Password string `json:"password"`
	SmtpHost string `json:"smtphost"`
	SmtpPort string `json:"smtpPort"`
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

		if r.URL.Query()["Defi"] != nil {
			num, _ := strconv.Atoi(r.URL.Query()["Defi"][0])
			data.DefiNumSelect = r.URL.Query()["Defi"][0]
			data.Participants = BDD.GetParticipant(num)
			if etu := r.URL.Query()["Etudiant"]; etu != nil {
				fmt.Println(etu)
				f, err := os.Open(config.Path_scripts + "script_" + etu[0] + "_" + data.DefiNumSelect + ".sh")
				if err != nil {
					data.File[0] = "erreur pour récupérer le script de l'étudiant"
				} else {
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						data.File = append(data.File, scanner.Text())
					}
				}
				if etat := r.URL.Query()["Etat"]; etat != nil {

					if etat[0] == "1" {
						BDD.SaveResultat(etu[0], num, 0, true)
					} else {
						BDD.SaveResultat(etu[0], num, 1, true)
					}
				}
			}

			if r.URL.Query()["getResult"] != nil {
				fileName := "resultat_" + strconv.Itoa(num) + ".csv"
				testeur.CreateCSV(fileName, num)
				w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
				w.Header().Set("Content-Type", "application/octet-stream")
				http.ServeFile(w, r, fileName)
				os.Remove(fileName)
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
				return
			}
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

		if r.URL.Query()["form"][0] == "sendMail" {

			etudiants := BDD.GetEtudiantsMail()
			nbDefis := len(BDD.GetDefis())

			file, err := os.Open("mailConf.json")
			if err != nil {
				fmt.Println(err)
			}
			byteValue, err := ioutil.ReadAll(file)
			if err != nil {
				fmt.Println(err)
			}
			var configSender SenderData
			err = json.Unmarshal(byteValue, &configSender)
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

			sendOk := sendMail(etudiants, nbDefis, configSender)
			if sendOk == false {
				fmt.Println("Erreur lors de l'envoi de mails")
			} else {
				fmt.Println("Mail envoyés !")
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
		file, fileHeader, err := r.FormFile("upload")
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
			typeTest := fileHeader.Header.Values("Content-Type")
			fmt.Println(typeTest)
			//application/zip , application/tar, text/plain; charset=utf-8

			//if dossier de test existe déjà, on le supprime
			pathTest := config.Path_jeu_de_tests + "test_defi_" + strconv.Itoa(num_defi_actuel)
			if testeur.Contains(config.Path_jeu_de_tests, "test_defi_"+strconv.Itoa(num_defi_actuel)) {
				os.RemoveAll(pathTest)
			}
			fichier, _ := os.Create(config.Path_jeu_de_tests + fileHeader.Filename) // remplacer handler.Filename par le nom et on le place où on veut
			defer fichier.Close()
			_, err = io.Copy(fichier, file)
			os.Chmod(fichier.Name(), 777)

			if typeTest[0] == "application/zip" {

				cmd := exec.Command("unzip", "-d",
					"test_defi_"+strconv.Itoa(num_defi_actuel),
					fileHeader.Filename)
				cmd.Dir = config.Path_jeu_de_tests
				cmd.Run()
				dosTest := testeur.GetFiles(pathTest)
				if len(dosTest) == 1 {
					os.Rename(pathTest+"/"+dosTest[0], config.Path_jeu_de_tests+"temp")
					os.RemoveAll(pathTest)
					os.Rename(config.Path_jeu_de_tests+"temp", pathTest)
				}
			} else if typeTest[0] == "application/x-tar" {
				cmd := exec.Command("tar", "tf", fileHeader.Filename)
				cmd.Dir = config.Path_jeu_de_tests
				output, _ := cmd.CombinedOutput()
				nomArchive := strings.Split(string(output), "\n")[0]
				cmd = exec.Command("tar", "xvf", fileHeader.Filename)
				cmd.Dir = config.Path_jeu_de_tests
				if err := cmd.Run(); err != nil {
					fmt.Println(err.Error())
				}
				os.Rename(config.Path_jeu_de_tests+nomArchive, pathTest)
			}
			os.Remove(fichier.Name())
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}
		script, err := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut
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

func sendMail(etudiants []BDD.EtudiantMail, nbDefis int, config SenderData) bool { // Authentication sur le serveur de mail

	auth := smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)

	for _, etu := range etudiants {

		etudiant := etu
		go func() {

			// adresse du destinataire
			to := []string{
				etudiant.Mail,
			}

			// En-tête du mail
			header := make(map[string]string)
			header["From"] = config.FromMail
			header["To"] = to[0]
			header["Subject"] = "Defis du lundi"
			header["MIME-Version"] = "1.0"
			header["Content-Type"] = "text/plain; charset= utf-8"
			header["Content-Transfer-Encoding"] = "base64"

			// Création du contenu du mail
			message := ""
			for champ, valeur := range header {
				message += fmt.Sprintf("%s : %s\r\n", champ, valeur)
			}

			body := "Résultats des défis du lundi\n\n" +
				"Bonjour " + etudiant.Prenom + " " + etudiant.Nom + "\n" +
				"A ce jour vous avez réalisé " + strconv.Itoa(len(etudiant.Defis)) +
				" défis sur " + strconv.Itoa(nbDefis) + "\n\n"

			nbDefisReussi := 0

			if len(etudiant.Defis) > 0 {
				for _, defi := range etudiant.Defis {
					defiStr := ""
					if defi.Etat == 1 {
						defiStr = defiStr + "Vous avez réussi "
						nbDefisReussi++
					} else {
						defiStr = defiStr + "Vous n'avez pas réussi "
					}
					defiStr = defiStr + "le défi n°" + strconv.Itoa(defi.Defi) + ", vous avez fait " + strconv.Itoa(defi.Tentative) + " tentatives\n"
					body = body + defiStr
				}
			} else {
				body = body + "Vous n'avez participé à aucun défis \n"
			}

			pointsBonus := 0.0

			if nbDefisReussi == 0 {
				pointsBonus = 0.0
			} else if nbDefisReussi <= 2 {
				pointsBonus = 0.1
			} else if nbDefisReussi <= 4 {
				pointsBonus = 0.25
			} else if nbDefisReussi <= 6 {
				pointsBonus = 0.5
			} else if nbDefisReussi <= 9 {
				pointsBonus = 1
			} else if nbDefisReussi >= 10 {
				pointsBonus = 2
			}

			pointsBonusStr := fmt.Sprintf("%0.2f", pointsBonus)

			body = body + "\nAinsi vous avez réussi " + strconv.Itoa(nbDefisReussi) + " défis ce qui donne un bonus de " + pointsBonusStr + " points sur la moyenne d'ISI\n"

			// encodage du contenu en UTF-8 pour que les caractères spéciaux s'affichent
			message += base64.StdEncoding.EncodeToString([]byte(body))

			// Envoi du mail
			err := smtp.SendMail(config.SmtpHost+":"+config.SmtpPort, auth, config.FromMail, to, []byte(message))
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	return true
}
