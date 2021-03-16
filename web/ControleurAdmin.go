package web

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aodin/date"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
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
	"time"
)

type data_pageAdmin struct { /* Données envoyée à la page admin */
	EtuSelect     string
	DefiSelect    config.Defi
	AdminInfo     config.Admin
	Etudiants     []config.Etudiant
	Res_etu       []config.ResBDD
	ListeDefis    []config.Defi
	File          []string
	DefiActuel    config.Defi
	JeuDeTestSent string
	Participants  []config.ParticipantDefi
	Correcteur    config.Etudiant
	Tricheurs     [][]string
	Logs          []string
	Log           []string
	LogDate       string
}

type SenderData struct { /* Structure utile pour l'envoi de mail */
	FromMail string `json:"fromMail"`
	Username string `json:"username"`
	Password string `json:"password"`
	SmtpHost string `json:"smtphost"`
	SmtpPort string `json:"smtpPort"`
}

type ResultMail struct { /* Structure de retour de l'envoi de mail */
	adress string
	send   bool
	erreur string
}

type Admin struct {
	Login    string
	Password string
}

func pageAdmin(w http.ResponseWriter, r *http.Request) {
	//Si il y a n'y a pas de token dans les cookies alors l'utilisateur est redirigé vers la page de login
	if token, err := r.Cookie("token"); err != nil || !BDD.TokenExiste(token.Value) {
		http.Redirect(w, r, "/loginAdmin", http.StatusFound)
		return
	}

	token, _ := r.Cookie("token")            //récupère le token du cookie
	login := BDD.GetNameByToken(token.Value) // récupère le login correspondant au token
	admin := BDD.GetAdmin(login)             // récupère les informations de l'étudiant grâce au login

	data := data_pageAdmin{
		AdminInfo:  admin,
		Etudiants:  BDD.GetEtudiants(),
		DefiActuel: BDD.GetDefiActuel(),
		ListeDefis: BDD.GetDefis(),
		Logs:       manipStockage.GetFiles(config.Path_log),
	}
	//if date actuelle > defi actel.datefin alors defiactuel.num = -1
	if data.DefiActuel.Num != -1 {
		if !date.Today().Within(date.NewRange(data.DefiActuel.Date_debut, data.DefiActuel.Date_fin)) {
			data.DefiActuel.Num = -1
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
				data.Log = []string{"erreur pour récupérer le fichier de log"}
			} else {
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					data.Log = append(data.Log, scanner.Text())
				}
			}
		}

		if r.URL.Query()["Defi"] != nil {
			num, _ := strconv.Atoi(r.URL.Query()["Defi"][0])
			data.DefiSelect = BDD.GetDefi(num)
			data.Correcteur = BDD.GetCorrecteur(num)
			fmt.Println("data.correcteur = ", data.Correcteur)
			data.Participants = BDD.GetParticipant(num)
			if etu := r.URL.Query()["Etudiant"]; etu != nil {
				fmt.Println(etu)
				f, err := os.Open(config.Path_scripts + "script_" + etu[0] + "_" + strconv.Itoa(data.DefiSelect.Num))
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
						BDD.SaveResultat(etu[0], num, 0, nil, true)
					} else {
						BDD.SaveResultat(etu[0], num, 1, nil, true)
					}
				}
			}

			if r.URL.Query()["getResult"] != nil {
				fileName := "resultat_" + strconv.Itoa(num) + ".csv"
				manipStockage.CreateCSV(fileName, num)
				w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
				w.Header().Set("Content-Type", "application/octet-stream")
				http.ServeFile(w, r, fileName)
				os.Remove(fileName)
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
				return
			}
			if r.URL.Query()["Correcteur"] != nil {
				BDD.GenerateCorrecteur(num)
				etudiant := BDD.GetCorrecteur(num)
				etudiantMail := config.EtudiantMail{Prenom: etudiant.Prenom, Nom: etudiant.Nom}
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
				resultMail := sendMailCorrecteur(etudiantMail, num, configSender)
				if resultMail.send == false {
					logs.WriteLog("Envoi de mail correcteur", "Erreur lors de l'envoi de mail du correcteur du défi "+strconv.Itoa(num)+" à l'adresse : "+etudiantMail.Mail())
				} else {
					logs.WriteLog("Envoi de mail correcteur", "envoi de mail du correcteur du défi "+strconv.Itoa(num)+" à l'adresse : "+etudiantMail.Mail())
				}
				http.Redirect(w, r, "/pageAdmin?Defi="+strconv.Itoa(num), http.StatusFound)
				return
			}
			if r.URL.Query()["getIdentique"] != nil {
				data.Tricheurs = manipStockage.GetTriche(num)
			}
		}

		if r.URL.Query()["logout"] != nil {
			fmt.Println("logout " + admin.Login)
			DeleteToken(admin.Login, time.Second*0)
			http.Redirect(w, r, "/loginAdmin", http.StatusFound)
			return
		}

		t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))

		if err := t.Execute(w, data); err != nil {
			log.Printf("error exec template : ", err)
		}
	}

	if r.Method == "POST" {

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

			resultatsEnvois := sendMailResults(etudiants, nbDefis, configSender)
			for _, res := range resultatsEnvois {
				if res.send == false {
					logs.WriteLog("Envoi de mails : ", "Erreur lors de l'envoi de mails à l'adresse : "+res.adress)
				}
			}

			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}

		// Permet de récupérer les résultats de tous les étudiants ainsi que leurs informations pour un défi donné
		if r.URL.Query()["form"][0] == "getResult" {
			num := r.FormValue("num")
			n, err := strconv.Atoi(num)
			if err != nil {
				fmt.Println(err.Error())
			}
			file_name := "resultat_" + num + ".csv"
			manipStockage.CreateCSV(file_name, n)
			w.Header().Set("Content-Disposition", "attachment; filename="+file_name)
			w.Header().Set("Content-Type", "application/octet-stream")
			http.ServeFile(w, r, file_name)
			os.Remove(file_name)
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}

		if r.URL.Query()["form"][0] == "DeleteDefi" {
			lastDefi := data.ListeDefis[0]
			os.Remove(config.Path_defis + "correction_" + strconv.Itoa(lastDefi.Num))
			err := os.RemoveAll(config.Path_jeu_de_tests + "test_defi_" + strconv.Itoa(lastDefi.Num))
			if err != nil {
				fmt.Println(err.Error())
			}
			BDD.DeleteLastDefi(lastDefi.Num)
			logs.WriteLog("Delete défi", "vous avez supprimer le défi N°"+strconv.Itoa(lastDefi.Num))
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return

		}

		r.ParseMultipartForm(10 << 20)

		file, fileHeader, errorFile := r.FormFile("upload")
		if errorFile == nil {
			defer file.Close()
		}

		defi_actuel := BDD.GetDefiActuel()
		num_defi_actuel := defi_actuel.Num
		path := ""

		if r.URL.Query()["form"][0] == "modify-defi" {
			numDefi, _ := strconv.Atoi(r.FormValue("defiSelectModif")) // Et le num du defi

			if r.FormValue("date_debut") != "" {
				fmt.Println("change date defi")
				logs.WriteLog("Admin", "modification de la date de rendu")
				debut, _ := date.Parse(r.FormValue("date_debut")) // On récupère les date modifiée
				fin, _ := date.Parse(r.FormValue("date_fin"))
				BDD.ModifyDefi(numDefi, debut, fin)
			}
			if errorFile == nil {
				logs.WriteLog("Admin", "modification du défi actuel")
				path = config.Path_defis + "correction_" + strconv.Itoa(num_defi_actuel)
				script, _ := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut
				defer script.Close()
				io.Copy(script, file)
				os.Chmod(path, 770)
			}
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}
		if r.URL.Query()["form"][0] == "defi" { // ajout d'un défi
			date_debut, _ := date.Parse(r.FormValue("date_debut"))
			date_fin, _ := date.Parse(r.FormValue("date_fin"))

			logs.WriteLog("Admin", "ajout d'un nouveau défi du "+date_debut.String()+" au "+date_fin.String())
			// ajouter a la table défis
			BDD.AddDefi(date_debut, date_fin)
			os.Mkdir(config.Path_jeu_de_tests+"test_defi_"+strconv.Itoa(num_defi_actuel+1), os.ModePerm)
			num_defi_actuel = num_defi_actuel + 1
			path = config.Path_defis + "correction_" + strconv.Itoa(num_defi_actuel)

			script, _ := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut
			defer script.Close()
			io.Copy(script, file)
			os.Chmod(path, 770)
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}
		if r.URL.Query()["form"][0] == "test" { // Pour upload un test
			num := r.FormValue("defiSelectTest")
			if num == "" {
				logs.WriteLog("upload test", "aucun numéro de défis n'a été spécifié")
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
				return
			}

			num2, _ := strconv.Atoi(num)
			defi := BDD.GetDefi(num2)
			typeArchive := fileHeader.Header.Values("Content-Type")
			fmt.Println(typeArchive)

			if typeArchive[0] != "application/zip" && typeArchive[0] != "application/x-tar" && typeArchive[0] != "application/tar" {
				logs.WriteLog("upload test", "format de l'upload incompatible")
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
				return
			}
			logs.WriteLog("Admin", "upload d'un test pour le défi n°"+num)
			if !defi.JeuDeTest {
				BDD.AddJeuDeTest(num2)
			} else if defi.Num == defi_actuel.Num {
				// Si on change le défi ACTUEL
				BDD.ResetEtatDefi(num2)
				/*
				 * (soit tous les étudiants, soit uniquement les étudiant ayant envoyé un script (enregistré dans Resultat)
				 * pour leur dire que le jeu de test a changé et que leur résultat est repassé à "non testé"
				 */
				etudiants := BDD.GetEtudiantsMail()

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

				resultatsEnvois := sendMailChange(etudiants, defi_actuel.Num, configSender)
				for _, res := range resultatsEnvois {
					if res.send == false {
						logs.WriteLog("Envoi de mails : ", "Erreur lors de l'envoi de mails à l'adresse : "+res.adress+" erreur : "+res.erreur)
					}
				}
			}
			//if dossier de test existe déjà, on le supprime
			pathTest := config.Path_jeu_de_tests + "test_defi_" + num
			if manipStockage.Contains(config.Path_jeu_de_tests, "test_defi_"+num) {
				os.RemoveAll(pathTest)
			}
			fichier, _ := os.Create(config.Path_jeu_de_tests + fileHeader.Filename) // remplacer handler.Filename par le nom et on le place où on veut
			defer fichier.Close()
			io.Copy(fichier, file)
			os.Chmod(fichier.Name(), 777)

			if typeArchive[0] == "application/zip" {
				cmd := exec.Command("unzip", "-d",
					"test_defi_"+strconv.Itoa(num_defi_actuel),
					fileHeader.Filename)
				cmd.Dir = config.Path_jeu_de_tests
				cmd.Run()
				dosTest := manipStockage.GetFiles(pathTest)
				if len(dosTest) == 1 {
					os.Rename(pathTest+"/"+dosTest[0], config.Path_jeu_de_tests+"temp")
					os.RemoveAll(pathTest)
					os.Rename(config.Path_jeu_de_tests+"temp", pathTest)
				}
			} else if typeArchive[0] == "application/x-tar" || typeArchive[0] == "application/tar" {
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

		if r.URL.Query()["form"][0] == "changeId" {
			login := r.FormValue("loginAd")
			password := r.FormValue("passwordAd")
			BDD.RegisterAdminString(login, password)
			http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			return
		}
	}
}

func sendMailResults(etudiants []config.EtudiantMail, nbDefis int, config SenderData) []ResultMail { // Authentication sur le serveur de mail

	auth := smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)
	c := make(chan ResultMail)
	var resultatsEnvois []ResultMail

	for _, etu := range etudiants {

		etudiant := etu
		go func() {

			// adresse du destinataire
			to := []string{
				etudiant.Mail(),
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
				c <- ResultMail{adress: etudiant.Mail(), send: false}
			} else {
				c <- ResultMail{adress: etudiant.Mail(), send: true}
			}
		}()
		resultatsEnvois = append(resultatsEnvois, <-c)
	}
	return resultatsEnvois
}

func sendMailCorrecteur(etudiant config.EtudiantMail, nbDefi int, config SenderData) ResultMail {
	// Authentication sur le serveur de mail

	auth := smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)

	to := []string{
		etudiant.Mail(),
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

	body := "défis du lundi\n\n" +
		"Bonjour " + etudiant.Prenom + " " + etudiant.Nom + "\n" +
		"A ce jour vous avez réussi le défi  " + strconv.Itoa(nbDefi) + ".\nVous avez été aléatoirement " +
		"nommé correcteur pour ce test parmi ceux ayant réussi le test\n" +
		"Vous devez donc envoyer un mail au professeur afin de lui remettre votre correction\n \nBonne journée"

	// encodage du contenu en UTF-8 pour que les caractères spéciaux s'affichent
	message += base64.StdEncoding.EncodeToString([]byte(body))

	// Envoi du mail
	err := smtp.SendMail(config.SmtpHost+":"+config.SmtpPort, auth, config.FromMail, to, []byte(message))
	if err != nil {
		return ResultMail{adress: etudiant.Mail(), send: false}
	} else {
		return ResultMail{adress: etudiant.Mail(), send: true}
	}
}

func sendMailChange(etudiants []config.EtudiantMail, nbDefi int, config SenderData) []ResultMail { // Authentication sur le serveur de mail

	auth := smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)
	c := make(chan ResultMail)
	var resultatsEnvois []ResultMail

	for _, etu := range etudiants {

		etudiant := etu
		go func() {

			// adresse du destinataire
			to := []string{
				etudiant.Mail(),
			}

			// En-tête du mail
			header := make(map[string]string)
			header["From"] = config.FromMail
			header["To"] = to[0]
			header["Subject"] = "Defis du lundi : Changement jeu de test"
			header["MIME-Version"] = "1.0"
			header["Content-Type"] = "text/plain; charset= utf-8"
			header["Content-Transfer-Encoding"] = "base64"

			// Création du contenu du mail
			message := ""
			for champ, valeur := range header {
				message += fmt.Sprintf("%s : %s\r\n", champ, valeur)
			}

			body := "Changement des jeux de test pour le défis n°" + strconv.Itoa(nbDefi) + "\n\n" +
				"Bonjour " + etudiant.Prenom + " " + etudiant.Nom + "\n" +
				"Les test de correction ont été changés, ainsi votre script n'est plus enregistré comme testé" +
				"et n'est peut être plus valide. \n" +
				"Veuillez le retester afin de valider le défis"

			// encodage du contenu en UTF-8 pour que les caractères spéciaux s'affichent
			message += base64.StdEncoding.EncodeToString([]byte(body))

			// Envoi du mail
			err := smtp.SendMail(config.SmtpHost+":"+config.SmtpPort, auth, config.FromMail, to, []byte(message))
			if err != nil {
				c <- ResultMail{adress: etudiant.Mail(), send: false, erreur: err.Error()}
			} else {
				c <- ResultMail{adress: etudiant.Mail(), send: true}
			}
		}()
		resultatsEnvois = append(resultatsEnvois, <-c)
	}
	return resultatsEnvois
}
