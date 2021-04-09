package web

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/testeur"
	"html/template"
	"io"
	"io/ioutil"
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
	DefiSelect    modele.Defi
	AdminInfo     modele.Admin
	Etudiants     []modele.Etudiant
	Res_etu       []modele.Resultat
	ListeDefis    []modele.Defi
	File          []string
	DefiActuel    modele.Defi
	JeuDeTestSent string
	Participants  []modele.ParticipantDefi
	Correcteur    modele.Etudiant
	Tricheurs     [][]string
	Logs          []string
	Log           []string
	LogDate       string
	Alert         string
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

/**
Main : traite toutes les requettes de la page Admin
*/
func pageAdmin(w http.ResponseWriter, r *http.Request) {
	//Si il y a n'y a pas de token dans les cookies alors l'utilisateur est redirigé vers la page de login
	if token, err := r.Cookie("token"); err != nil || !DAO.TokenExiste(token.Value) {
		http.Redirect(w, r, "/loginAdmin", http.StatusFound)
		return
	}

	token, _ := r.Cookie("token")             //récupère le token du cookie
	login := DAO.GetLoginByToken(token.Value) // récupère le login correspondant au token
	admin := DAO.GetAdmin(login)              // récupère les informations de l'étudiant grâce au login

	data := data_pageAdmin{
		AdminInfo:  admin,
		Etudiants:  DAO.GetEtudiants(),
		DefiActuel: DAO.GetDefiActuel(),
		ListeDefis: DAO.GetDefis(),
		Logs:       manipStockage.GetFiles(modele.PathLog),
		Alert:      "",
	}

	fmt.Println(r.URL.String())
	fmt.Println("methode : " + r.Method)
	if r.Method == "GET" {

		// Permet d'afficher les logs d'une date précise
		if r.URL.Query()["Log"] != nil {
			log := r.URL.Query()["Log"][0]
			data.LogDate = log
			f, err := os.Open(modele.PathLog + log)
			if err != nil {
				data.Log = []string{"erreur pour récupérer le fichier de log"}
				data.Alert = "Erreur pour récupérer le fichier de log" + log
			} else {
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					data.Log = append(data.Log, scanner.Text())
				}
			}
		}

		if r.URL.Query()["Defi"] != nil {
			num, _ := strconv.Atoi(r.URL.Query()["Defi"][0])
			data.DefiSelect = DAO.GetDefi(num)
			data.Correcteur = DAO.GetCorrecteur(num)
			data.Participants = DAO.GetParticipants(num)
			if etu := r.URL.Query()["Etudiant"]; etu != nil {
				f, err := os.Open(modele.PathScripts + "script_" + etu[0] + "_" + strconv.Itoa(data.DefiSelect.Num))
				if err != nil {
					data.File[0] = "erreur pour récupérer le script_E197051L_1 de l'étudiant"
				} else {
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						data.File = append(data.File, scanner.Text())
					}
				}
				if etat := r.URL.Query()["Etat"]; etat != nil {
					if etat[0] == "1" {
						DAO.SaveResultat(etu[0], num, 0, nil, true)
					} else {
						DAO.SaveResultat(etu[0], num, 1, nil, true)
					}
					http.Redirect(w, r, "/pageAdmin?Defi="+strconv.Itoa(num), http.StatusFound)
					return
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

			// Choisi l'étudiant correcteur et lui envoi un mail
			if r.URL.Query()["Correcteur"] != nil && DAO.GetCorrecteur(num).Login == "" {
				DAO.GenerateCorrecteur(num)
				etudiant := DAO.GetCorrecteur(num)
				if etudiant.Login == "" {
					data.Alert = "Aucun correcteur n'a été trouvé"
					logs.WriteLog("GetCorrecteur", data.Alert)
				} else {
					etudiantMail := modele.EtudiantMail{Prenom: etudiant.Prenom, Nom: etudiant.Nom}
					file, err := os.Open("mailConf.json")
					if err != nil {
						logs.WriteLog("Envoie de mail correcteur", "Erreur mailConf.json est introuvable")
					}
					byteValue, _ := ioutil.ReadAll(file)
					var configSender SenderData
					err = json.Unmarshal(byteValue, &configSender)
					if err != nil {
						logs.WriteLog("Envoie de mail correcteur", "Erreur unmarshal mailConf.json")
					}
					defer file.Close()
					resultMail := sendMailCorrecteur(etudiantMail, num, configSender)
					if resultMail.send == false {
						data.Alert = "Erreur lors de l'envoi de mail du correcteur du défi " + strconv.Itoa(num) + " à l'adresse : " + etudiantMail.Mail()
						logs.WriteLog("Envoi de mail correcteur", data.Alert)
					} else {
						logs.WriteLog("Envoi de mail correcteur", "envoi de mail du correcteur du défi "+strconv.Itoa(num)+" à l'adresse : "+etudiantMail.Mail())
					}
					data.Correcteur = DAO.GetCorrecteur(num)

				}
			}
			if r.URL.Query()["getIdentique"] != nil {
				data.Tricheurs = manipStockage.GetTriche(num)
			}
		}

		// Lors de la deconnexion
		if r.URL.Query()["logout"] != nil {
			logs.WriteLog("Page admin", "deconnexion de "+admin.Login)
			DeleteToken(admin.Login, time.Second*0)              // Le token est supprimé
			http.Redirect(w, r, "/loginAdmin", http.StatusFound) // On retourne à la page de connexion (celle de l'admin)
			return
		}

		t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))

		if err := t.Execute(w, data); err != nil {
			logs.WriteLog("Erreur execution template", err.Error())
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}

	if r.Method == "POST" {

		// Envoi de mail pour la progressions des étudiants
		if r.URL.Query()["form"][0] == "sendMail" {

			// On récupère les informations de la BDD
			etudiants := DAO.GetEtudiantsMail()
			nbDefis := len(DAO.GetDefis())

			// On lit dans le fichier de configuration les informations pour l'envoi de mails
			file, err := os.Open("mailConf.json")
			if err != nil {
				logs.WriteLog("Envoi de mails progression : ", err.Error())
			}
			byteValue, err := ioutil.ReadAll(file)
			if err != nil {
				logs.WriteLog("Envoi de mails progression : ", err.Error())
			}
			var configSender SenderData
			err = json.Unmarshal(byteValue, &configSender)
			if err != nil {
				logs.WriteLog("Envoi de mails progression : ", err.Error())
			}
			defer file.Close()

			// On appelle la fonction qui se charge de l'envoi de mails
			resultatsEnvois := sendMailResults(etudiants, nbDefis, configSender)
			fmt.Println(resultatsEnvois)
			for _, res := range resultatsEnvois {
				if res.send == false {
					logs.WriteLog("Envoi de mails : ", "Erreur lors de l'envoi de mails à l'adresse : "+res.adress)
				}
			}
			data.Alert = "L'envoie de mail a été effectué"
		}

		// Permet de récupérer les résultats de tous les étudiants ainsi que leurs informations pour un défi donné
		if r.URL.Query()["form"][0] == "getResult" {
			num := r.FormValue("num")
			n, err := strconv.Atoi(num)
			if err != nil {
				data.Alert = "Le numéro de défi entré n'est pas valide"
				logs.WriteLog("getResult CSV", data.Alert)
			} else {
				file_name := "resultat_" + num + ".csv"
				manipStockage.CreateCSV(file_name, n)
				w.Header().Set("Content-Disposition", "attachment; filename="+file_name)
				w.Header().Set("Content-Type", "application/octet-stream")
				http.ServeFile(w, r, file_name)
				os.Remove(file_name)

			}
		}

		if r.URL.Query()["form"][0] == "DeleteDefi" {
			if len(data.ListeDefis) > 0 {
				lastDefi := data.ListeDefis[len(data.ListeDefis)-1]
				os.Remove(modele.PathDefis + "correction_" + strconv.Itoa(lastDefi.Num))
				err := os.RemoveAll(modele.PathJeuDeTests + "test_defi_" + strconv.Itoa(lastDefi.Num))
				if err != nil {
					fmt.Println(err.Error())
				}
				DAO.DeleteLastDefi(lastDefi.Num)
				data.Alert = "Vsou avez supprimé le défi N°" + strconv.Itoa(lastDefi.Num)
			} else {
				data.Alert = "Vous ne pouvez pas supprimer un défi si la liste est vide"
			}
			logs.WriteLog("Delete défi", data.Alert)

		}

		r.ParseMultipartForm(10 << 20)

		file, fileHeader, errorFile := r.FormFile("upload")
		if errorFile == nil {
			defer file.Close()
		}

		defi_actuel := DAO.GetDefiActuel()
		num_defi_actuel := defi_actuel.Num
		path := ""

		if r.URL.Query()["form"][0] == "modify-defi" {
			numDefi, _ := strconv.Atoi(r.FormValue("defiSelectModif")) // Et le num du defi
			fmt.Println()
			if r.FormValue("date_debut") != defi_actuel.DateDebutString() || r.FormValue("date_fin") != defi_actuel.DateFinString() ||
				r.FormValue("time_debut") != defi_actuel.TimeDebutString() || r.FormValue("time_fin") != defi_actuel.TimeFinString() {

				layout := "2006-01-02T15:04:05.000Z"
				str := fmt.Sprintf("%sT%sZ", r.FormValue("date_debut"), r.FormValue("time_debut")+":00.000")
				t_debut, err := time.Parse(layout, str)
				if err != nil {
					data.Alert = "La date de départ est mal formée"
				}
				str = fmt.Sprintf("%sT%sZ", r.FormValue("date_fin"), r.FormValue("time_fin")+":00.000")
				t_fin, err := time.Parse(layout, str)
				if err != nil {
					data.Alert = "La date de fin est mal formée"
				}
				if data.Alert == "" {
					if t_fin.Sub(t_debut) < 0 { // t_fin est avant t_debut
						data.Alert = "La date de fin doit être après la date de début"
					} else {
						bool, err := verifDate(numDefi, t_debut, t_fin)
						if bool {
							logs.WriteLog("Admin", "modification de la date de rendu du défi "+strconv.Itoa(numDefi))
							DAO.ModifyDefi(numDefi, t_debut, t_fin)
						} else {
							data.Alert = err
							logs.WriteLog("Admin modification d'un défi", data.Alert)
						}
					}
				} else {
					logs.WriteLog("Admin modif defi", data.Alert)
				}
			}
			if errorFile == nil {
				logs.WriteLog("Admin", "modification du défi actuel")
				path = modele.PathDefis + "correction_" + strconv.Itoa(num_defi_actuel)
				script, _ := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut
				defer script.Close()
				io.Copy(script, file)
				os.Chmod(path, 770)
			}

		}
		if r.URL.Query()["form"][0] == "defi" { // ajout d'un défi
			layout := "2006-01-02T15:04:05.000Z"
			str := fmt.Sprintf("%sT%sZ", r.FormValue("date_debut"), r.FormValue("time_debut")+":00.000")
			t_debut, err := time.Parse(layout, str)
			if err != nil {
				data.Alert = "La date de départ est mal formée"
			}
			str = fmt.Sprintf("%sT%sZ", r.FormValue("date_fin"), r.FormValue("time_fin")+":00.000")
			t_fin, err := time.Parse(layout, str)
			if err != nil {
				data.Alert = "La date de fin est mal formée"
			}
			if data.Alert == "" {
				if t_fin.Sub(t_debut) < 0 { // t_fin est après t_debut
					data.Alert = "La date de fin doit être après la date de début"
				} else { // Si tout est bon, on ajoute dans la bdd et dans les ressource le script de correction et le dossier de test
					bool, err := verifDate(num_defi_actuel, t_debut, t_fin)
					if bool {
						logs.WriteLog("Admin", "ajout d'un nouveau défi du "+t_debut.String()+" au "+t_fin.String())
						DAO.AddDefi(t_debut, t_fin)
						num_defi_actuel++
						os.Mkdir(modele.PathJeuDeTests+"test_defi_"+strconv.Itoa(num_defi_actuel), os.ModePerm)
						path = modele.PathDefis + "correction_" + strconv.Itoa(num_defi_actuel)
						script, _ := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut
						io.Copy(script, file)
						os.Chmod(path, 770)
						script.Close()
					} else {
						data.Alert = err
						logs.WriteLog("Admin ajout d'un défi", data.Alert)
					}
				}
			}
		}
		if r.URL.Query()["form"][0] == "test" { // Pour upload un test
			num, _ := strconv.Atoi(r.FormValue("defiSelectTest"))
			defi := DAO.GetDefi(num)
			typeArchive := fileHeader.Header.Values("Content-Type")
			fmt.Println(typeArchive)

			if typeArchive[0] != "application/zip" && typeArchive[0] != "application/x-tar" && typeArchive[0] != "application/tar" {
				data.Alert = "Le format " + typeArchive[0] + " n'est pas supporté"
				logs.WriteLog("upload test", data.Alert)

				t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))
				if err := t.Execute(w, data); err != nil {
					logs.WriteLog("Erreur execution template", err.Error())
				}
			}

			logs.WriteLog("Admin", "upload d'un test pour le défi n°"+strconv.Itoa(num))
			if !defi.JeuDeTest {
				DAO.AddJeuDeTest(num)
			} else if defi.Num == defi_actuel.Num {
				// Si on change le défi ACTUEL
				//Récupère les étudiants ayant réussi le test avant que le jeu de test change
				etudiantsReussi := DAO.GetResultatsByEtat(num, 1)
				for _, etu := range etudiantsReussi { //retest le scripts de ces étudiants
					testeur.Test(etu.Login)
				}
				etudiantsFailed := DAO.GetResultatsByEtat(num, 0)
				etudiantsFailed = append(etudiantsFailed, DAO.GetResultatsByEtat(num, -1)...)
				//on récupère un string contenant tous les logins des étudiants qui sont passés de l'état réussi à échoué après avoir changé le jeu de test
				loginToSendMail := ""
				for _, etuFail := range etudiantsFailed {
					for _, etuSucess := range etudiantsReussi {
						if etuFail.Login == etuSucess.Login {
							loginToSendMail = loginToSendMail + " " + etuFail.Login
						}
					}
				}
				etuToSendMail := make([]modele.EtudiantMail, 0)
				for _, etu := range DAO.GetEtudiantsMail() { //Pour récupérer que les mails des étudiants à qui on veut envoyer un mail
					if strings.Contains(loginToSendMail, etu.Login) {
						etuToSendMail = append(etuToSendMail, etu)
					}
				}

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

				resultatsEnvois := sendMailChange(etuToSendMail, defi_actuel.Num, configSender)
				for _, res := range resultatsEnvois {
					if res.send == false {
						logs.WriteLog("Envoi de mails : ", "Erreur lors de l'envoi de mails à l'adresse : "+res.adress+" erreur : "+res.erreur)
					}
				}
			}
			//if dossier de test existe déjà, on le supprime
			pathTest := modele.PathJeuDeTests + "test_defi_" + strconv.Itoa(num)
			if manipStockage.Contains(modele.PathJeuDeTests, "test_defi_"+strconv.Itoa(num)) {
				os.RemoveAll(pathTest)
			}
			fichier, _ := os.Create(modele.PathJeuDeTests + fileHeader.Filename) // remplacer handler.Filename par le nom et on le place où on veut
			defer fichier.Close()
			io.Copy(fichier, file)
			os.Chmod(fichier.Name(), 777)

			if typeArchive[0] == "application/zip" {
				cmd := exec.Command("unzip", "-d",
					"test_defi_"+strconv.Itoa(num),
					fileHeader.Filename)
				cmd.Dir = modele.PathJeuDeTests
				cmd.Run()
				dosTest := manipStockage.GetFiles(pathTest)
				if len(dosTest) == 1 {
					os.Rename(pathTest+"/"+dosTest[0], modele.PathJeuDeTests+"temp")
					os.RemoveAll(pathTest)
					os.Rename(modele.PathJeuDeTests+"temp", pathTest)
				}
			} else if typeArchive[0] == "application/x-tar" || typeArchive[0] == "application/tar" {
				cmd := exec.Command("tar", "tf", fileHeader.Filename)
				cmd.Dir = modele.PathJeuDeTests
				output, _ := cmd.CombinedOutput()
				nomArchive := strings.Split(string(output), "\n")[0]
				cmd = exec.Command("tar", "xvf", fileHeader.Filename)
				cmd.Dir = modele.PathJeuDeTests
				if err := cmd.Run(); err != nil {
					fmt.Println(err.Error())
				}
				os.Rename(modele.PathJeuDeTests+nomArchive, pathTest)
			}
			os.Remove(fichier.Name())
		}

		// Ajoute un nouveau couple login:passwd dans la table Admin
		if r.URL.Query()["form"][0] == "changeId" {
			login := r.FormValue("loginAd")
			password := r.FormValue("passwordAd")
			if DAO.RegisterAdmin(login, password) {
				data.Alert = "changement d'id admin effectué"
			} else {
				data.Alert = "Erreur lors du changement d'id (voir logs)"
			}
			logs.WriteLog("Admin changement d'ID", data.Alert)
		}

		// Changer la configuation de l'envoi de mail (mailConf.json)
		if r.URL.Query()["form"][0] == "changeConfig" {

			/* Données du form */
			mail := r.FormValue("mailConf")
			username := r.FormValue("usernameConf")
			password := r.FormValue("passwordConf")
			host := r.FormValue("hostConf")
			port := r.FormValue("portConf")

			//Fichier de config
			err := os.Remove(modele.PathRoot + "mailConf.json") // On le suppr pour être sûr
			if err != nil {
				logs.WriteLog("Changement MailConf", "Pas de fichier mail.Conf.json")
			}
			fConf, err := os.OpenFile(modele.PathRoot+"mailConf.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend) // On l'ouvre
			if err != nil {
				data.Alert = "Erreur pour récupérer le fichier de config de mail"
			} else {
				// On écrit dedans sous forme d'un Json ce qui est utile
				data.Alert = "mailConf.json a été mis à jour"
				newConfString := "{\n  \"fromMail\" : \" " + mail + "\",\n  \"username\" : \"" + username + "\",\n  \"password\" : \"" + password + "\",\n  \"smtpHost\" : \"" + host + "\",\n  \"smtpPort\" : \"" + port + "\"\n}"
				_, err := fConf.Write([]byte(newConfString))
				if err != nil {
					data.Alert = "Erreur écriture dans le fichier mailConf.json"
				}
			}
			logs.WriteLog("Changement Mail config", data.Alert)
			fConf.Close()
		}

		data.ListeDefis = DAO.GetDefis()
		data.DefiActuel = DAO.GetDefiActuel()
		data.Etudiants = DAO.GetEtudiants()

		t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))
		if err := t.Execute(w, data); err != nil {
			logs.WriteLog("Erreur execution template", err.Error())
		}
	}
}

func verifDate(numDefi int, t_debut, t_fin time.Time) (bool, string) {
	ListeDefis := DAO.GetDefis()
	b := true
	var err string
	for _, defi := range ListeDefis {
		if defi.Num != numDefi {
			if t_debut.Sub(defi.DateDebut) > 0 && t_debut.Sub(defi.DateFin) < 0 {
				b = false
				err = "la date de début est pendant la période du défi " + strconv.Itoa(defi.Num)
			} else if t_fin.Sub(defi.DateDebut) > 0 && t_debut.Sub(defi.DateFin) < 0 {
				b = false
				err = "la date de fin est pendant la période du défi " + strconv.Itoa(defi.Num)
			} else if t_debut.Sub(defi.DateDebut) < 0 && t_fin.Sub(defi.DateFin) > 0 {
				b = false
				err = "Il y a le défi " + strconv.Itoa(defi.Num) + "entre la date de début et de fin"
			}
		}
	}
	return b, err
}

/*
	Fonction qui permet d'envoyer à tout les étudiants leurs progressions
*/
func sendMailResults(etudiants []modele.EtudiantMail, nbDefis int, config SenderData) []ResultMail {
	// Authentication sur le serveur de mail
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)
	// Canal pour recevoir les informations depuis les goroutines
	c := make(chan ResultMail)
	// Collections pour stocker des informations sur l'envoi de mails
	var resultatsEnvois []ResultMail

	// On boucle sur chaque étudiant
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
				logs.WriteLog("Envoi de mails progression", err.Error())
			} else {
				c <- ResultMail{adress: etudiant.Mail(), send: true}
			}
		}()
	}
	i := 0
	for i < len(etudiants) {
		// On récupère via le canal le résultat de l'envoi pour chaque étudiant
		resultatsEnvois = append(resultatsEnvois, <-c)
		i++
	}
	return resultatsEnvois
}

func sendMailCorrecteur(etudiant modele.EtudiantMail, nbDefi int, config SenderData) ResultMail {
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

func sendMailChange(etudiants []modele.EtudiantMail, nbDefi int, config SenderData) []ResultMail { // Authentication sur le serveur de mail

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
				"Les test de correction ont été changés, ainsi votre script_E197051L_1 n'est plus enregistré comme testé" +
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
