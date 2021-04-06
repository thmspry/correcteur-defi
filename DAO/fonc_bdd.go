package DAO

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var db, _ = sql.Open("sqlite3", "./DAO/database.db")
var m sync.Mutex

/**
Fonction qui initialise les tables vides
*/
func InitDAO() {

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS Etudiant (" +
		"login TEXT PRIMARY KEY, " +
		"password TEXT NOT NULL, " +
		"prenom TEXT NOT NULL," +
		"nom TEXT NOT NULL," +
		"correcteur BOOLEAN NOT NULL," +
		"resDefiActuel TEXT" +
		");")
	if err != nil {
		fmt.Println("prblm table Etudiant" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Defis (" +
		"numero INTEGER PRIMARY KEY AUTOINCREMENT," +
		"dateDebut TEXT NOT NULL," +
		"dateFin TEXT NOT NULL," +
		"jeuDeTest BOOL NOT NULL," +
		"correcteur TEXT," +
		"FOREIGN KEY (correcteur) REFERENCES Etudiant(login)" +
		")")
	if err != nil {
		fmt.Println("Erreur création table Defis " + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Resultat (" +
		"login TEXT NOT NULL," +
		"defi INTEGER NOT NULL," +
		"etat INTEGER NOT NULL," + // 3 états possibles : 1 (réussi), 0 (non réussi), -1 (non testé)
		"tentative INTEGER NOT NULL," + // Nombre de tentative au test
		"FOREIGN KEY (login) REFERENCES Etudiant(login)" +
		"FOREIGN KEY (defi) REFERENCES Defis(numero)" +
		")")
	if err != nil {
		fmt.Println("prblm table ResTest" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Token (" +
		"login TEXT NOT NULL PRIMARY KEY ," +
		"token TEXT NOT NULL," +
		"FOREIGN KEY(login) REFERENCES Etudiant(login)" +
		")")
	if err != nil {
		fmt.Println("Erreur dans la table Token" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Administrateur (" +
		"login TEXT NOT NULL PRIMARY KEY ," +
		"password TEXT NOT NULL" +
		")")
	if err != nil {
		fmt.Println("Erreur dans la table Administrateur" + err.Error())
	}
	stmt.Exec()

	stmt.Close()

	admin := modele.Admin{
		Login:    "admin",
		Password: "admin",
	}
	RegisterAdmin(admin)
}

/**
Enregistre un étudiant dans la table Etudiant
*/
func Register(etu modele.Etudiant) bool {
	m.Lock()
	stmt, err := db.Prepare("INSERT INTO Etudiant(login,password,prenom,nom,correcteur) values(?,?,?,?,?)")
	if err != nil {
		logs.WriteLog("DAO register étudiant : ", err.Error())
	}
	_, err = stmt.Exec(etu.Login, etu.Password, etu.Prenom, etu.Nom, false)
	if err != nil {
		logs.WriteLog("DAO register étudiant : ", err.Error())
		m.Unlock()
		return false
	}
	logs.WriteLog("Register étudiant", etu.Login+" est enregistré")
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("Register étudiant : ", err.Error())
	}
	m.Unlock()
	return true
}

/**
Enregistre un admin dans la table Administrateur
*/
func RegisterAdmin(admin modele.Admin) bool {
	m.Lock()
	stmt, err := db.Prepare("INSERT INTO Administrateur values(?,?)")
	if err != nil {
		logs.WriteLog("DAO register admin : ", err.Error())
	}
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(admin.Password), 14)
	_, err = stmt.Exec(admin.Login, passwordHashed)
	if err != nil {
		logs.WriteLog("DAO register admin : ", err.Error())
		m.Unlock()
		return false
	}
	fmt.Println("l'admin de login : " + admin.Login + " a été enregistré dans la bdd\n")
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("DAO register admin : ", err.Error())
	}
	m.Unlock()
	return true
}

/**

 */
func RegisterAdminString(login string, password string) bool {
	admin := modele.Admin{
		Login:    login,
		Password: password,
	}
	return RegisterAdmin(admin)
}

/**
Vérifie que le couple login,password existe dans la table Etudiant
*/
func LoginCorrect(id string, password string) bool {
	m.Lock()
	var passwordHashed string
	row := db.QueryRow("SELECT password FROM Etudiant WHERE login = $1", id)
	if row == nil { // pas de compte avec ce login
		return false
	}
	err := row.Scan(&passwordHashed) // cast/parse du res de la requète en string dans passwordHashed
	if err != nil {
		logs.WriteLog("DAO.LoginCorrect", err.Error())
	}
	errCompare := bcrypt.CompareHashAndPassword([]byte(passwordHashed), []byte(password)) // comparaison du hashé et du clair
	m.Unlock()
	return errCompare == nil // si nil -> ça match, sinon non
}

func IsLoginUsed(id string) bool {
	var pseudo string
	row := db.QueryRow("SELECT login FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&pseudo)
	if err != nil {
		return false
	}
	return true
}

/**
vérifie que le couple login,password existe dans la table Administrateur
*/
func LoginCorrectAdmin(id string, password string) bool {
	m.Lock()
	var passwordHashed string
	row := db.QueryRow("SELECT password FROM administrateur WHERE login = $1", id)
	errScan := row.Scan(&passwordHashed) // cast/parse du res de la requète en string dans passwordHashed
	if errScan != nil {
		logs.WriteLog(id, "login admin inconnu")
	}
	errCompare := bcrypt.CompareHashAndPassword([]byte(passwordHashed), []byte(password)) // comparaison du hashé et du clair
	m.Unlock()
	return errCompare == nil // si nil -> ça match, sinon non
}

/**
récupère les informations personnelles d'un étudiant
*/
func GetEtudiant(id string) modele.Etudiant {
	var etu modele.Etudiant
	row := db.QueryRow("SELECT * FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Correcteur, &etu.ResDefiActuel)
	if err != nil && etu.ResDefiActuel != nil {
		logs.WriteLog("DAO.GetEtudiant", err.Error())
	}
	return etu
}

/**
récupère les informations d'un admin
*/
func GetAdmin(id string) modele.Admin {
	var admin modele.Admin
	row := db.QueryRow("SELECT * FROM Administrateur WHERE login = $1", id)
	err := row.Scan(&admin.Login, &admin.Password)

	if err != nil {
		logs.WriteLog("DAO GetAdmin "+id+" : ", err.Error())
	}
	return admin
}

/**
@GetLoginByToken récupère le login par le token
*/
func GetLoginByToken(token string) string {
	var login string
	row := db.QueryRow("SELECT * FROM token WHERE token = $1", token)
	err := row.Scan(&login, &token)
	if err != nil {
		logs.WriteLog("DAO GetLoginByToken "+token+" : ", err.Error())
	}
	return login
}

/**
@InsertToken, insert un token pour un étudiant, détruis tous les tokens de cet étudiant avant ça
*/
func InsertToken(login string, token string) {
	m.Lock()
	stmt, _ := db.Prepare("DELETE FROM Token where login = ?")
	_, err := stmt.Exec(login)
	if err != nil {
		logs.WriteLog("DAO InsertToken "+login, err.Error())
	}
	stmt, err = db.Prepare("INSERT INTO Token values(?,?)")
	if err != nil {
		logs.WriteLog("DAO InsertToken "+login, err.Error())
	}
	_, err = stmt.Exec(login, token)
	if err != nil {
		logs.WriteLog("DAO InsertToken "+login, err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("DAO InsertToken "+login, err.Error())
	}
	m.Unlock()
}

/**
@DeleteToken, supprime le(s) token(s) correspondant au login donné
@login login de l'étudiant ayant le token
*/
func DeleteToken(login string) {
	m.Lock()
	stmt, err := db.Prepare("DELETE FROM token WHERE login = ?")
	if err != nil {
		logs.WriteLog("DAO.DeleteToken", err.Error())
	}
	_, err = stmt.Exec(login)
	if err != nil {
		logs.WriteLog("DAO DeleteToken", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("DAO.DeleteToken", err.Error())
	}
	m.Unlock()
}

func TokenExiste(token string) bool {
	var (
		log string
		tok string
	)
	row := db.QueryRow("SELECT * FROM token WHERE token = $1", token)
	err := row.Scan(&log, &tok)
	if err != nil {
		return false
	}
	return true
}

func TokenRole(token string) string {
	var (
		login string
	)
	row := db.QueryRow("SELECT login FROM token WHERE token = $1", token)
	err := row.Scan(&login)
	if err != nil {
		return ""
	}

	var nb int
	row = db.QueryRow("SELECT  count(*) FROM etudiant WHERE login = $1", login)
	err = row.Scan(&nb)
	if err != nil {

	}
	if nb == 1 {
		return "etudiants"
	}

	row = db.QueryRow("SELECT  count(*) FROM administrateur WHERE login = $1", login)
	err = row.Scan(&nb)
	if nb == 1 {
		return "administrateur"
	}

	logs.WriteLog("DAO TokenRole", "Le login associé n'est pas dans la table administrateur ou étudiant")
	return ""
}

/**
@SaveResultat permet d'enregistré le résultat obtenu d'un étudiant pour un défi
@login le login de l'étudiant
@numDefi le numéro du défi concerné
@etat l'état du test effectué (1 : réussi, 0 : raté)
@resultat le résultat obtenu par le test
@admin == true : la fonction a été lancé par l'admin
	   == false : la fonction a été lancé lorsque l'étudiant a testé son script (alors on augmente le compteur de tentative)
*/
func SaveResultat(login string, numDefi int, etat int, resultat []modele.ResultatTest, admin bool) {
	m.Lock()

	if resultat != nil { //ajout du résultat obtenu à l'étudiant
		resJson, _ := json.Marshal(resultat)
		stmt, err := db.Prepare("UPDATE Etudiant SET resDefiActuel = ? WHERE login = ?")
		if err != nil {
			logs.WriteLog("DAO.SaveResultat", err.Error())
		}
		_, err = stmt.Exec(string(resJson), login)
		if err != nil {
			logs.WriteLog("DAO DeleteToken", err.Error())
		}
		err = stmt.Close()
	}

	var res modele.Resultat
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", login, numDefi)

	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		//si err diff nil, cela veut dire qu'il n'a pas réussi à scan car il n'y a pas de ligne dans row
		stmt, _ := db.Prepare("INSERT INTO Resultat values(?,?,?,?)")
		_, err = stmt.Exec(login, numDefi, etat, 1)
		err = stmt.Close()
		if err != nil {
			logs.WriteLog("DAO SaveResultat", err.Error())
		}
	} else {
		stmt, err := db.Prepare("UPDATE Resultat SET etat = ?, tentative = ? WHERE login = ? AND defi = ?")
		if err != nil {
			logs.WriteLog("DAO SaveResultats", err.Error())
		} else {
			if admin {
				_, err = stmt.Exec(etat, res.Tentative, res.Login, res.Defi)
				if err != nil {
					logs.WriteLog("DAO SaveResultats", err.Error())
				}

			} else {
				_, err = stmt.Exec(etat, res.Tentative+1, res.Login, res.Defi)
				if err != nil {
					logs.WriteLog("DAO SaveResultats", err.Error())
				}
			}
			err = stmt.Close()
			if err != nil {
				logs.WriteLog("DAO SaveResultats", err.Error())
			}
		}
	}
	m.Unlock()
}

/**
@GetResultatTest permet de récupérer le dernier résultat obtenu par un étudiant lorsqu'il test son script pour le défi actuel
@login le login de l'étudiant
*/
func GetResultatTest(login string) []modele.ResultatTest {
	var (
		query string
		res   []modele.ResultatTest
	)
	row := db.QueryRow("SELECT resDefiActuel FROM Etudiant WHERE login = $1", login)
	if err := row.Scan(&query); err != nil && query != "" {
		logs.WriteLog("DAO.GetResultatTest", err.Error())
	}
	json.Unmarshal([]byte(query), &res)
	return res
}

/**
@GetEtudiants retourne la liste des étudiants enregistrés dans la database
*/
func GetEtudiants() []modele.Etudiant {
	var etu modele.Etudiant
	etudiants := make([]modele.Etudiant, 0)
	row, err := db.Query("SELECT * FROM Etudiant")
	if err != nil {
		logs.WriteLog("DAO GetEtudiants", err.Error())
	}
	for row.Next() {
		err = row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Correcteur, &etu.ResDefiActuel)
		if err != nil && etu.ResDefiActuel != nil {
			logs.WriteLog("DAO GetEtudiants", err.Error())
		}
		etudiants = append(etudiants, etu)
	}
	row.Close()
	return etudiants
}

/**
@GetResult retourne le résultat d'un étudiant à un défi
@login login de l'étudiant
@defi numéro du défi concerné
*/
func GetResult(login string, defi int) modele.Resultat {
	var res modele.Resultat
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", login, defi)
	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		logs.WriteLog("DAO.GetResult", err.Error())

	}
	return res
}

/**
@AddDefi ajoute un défi à la database
@dateD date de début du défi
@dateF date de fin du défi
*/
func AddDefi(dateD time.Time, dateF time.Time) {
	m.Lock()
	tDeb, err := dateD.MarshalText()
	tFin, err := dateF.MarshalText()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt, err := db.Prepare("INSERT INTO Defis(dateDebut,dateFin, jeuDeTest) values(?,?,?)")
	if err != nil {
		logs.WriteLog("DAO AddeDefi", err.Error())
	} else {
		_, err = stmt.Exec(string(tDeb), string(tFin), false)
		if err != nil {
			logs.WriteLog("DAO AddDefi", err.Error())
		}
		stmt.Close()
	}
	m.Unlock()
}

/**
@ModifyDefi modifie les dates d'un défi déjà enregistré
@num numéro du défi qu'on modifie
@dateD nouvelle date de début du défi
@dateF nouvelle date de fin du défi
*/
func ModifyDefi(num int, dateD time.Time, dateF time.Time) {
	tDeb, err := dateD.MarshalText()
	tFin, err := dateF.MarshalText()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt, err := db.Prepare("UPDATE Defis SET dateDebut = ?, dateFin = ? where numero = ?")
	if err != nil {
		logs.WriteLog("DAO modify defi", err.Error())
	}
	if _, err := stmt.Exec(string(tDeb), string(tFin), num); err != nil {
		logs.WriteLog("DAO.ModifyDefi", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("DAO modify defi", err.Error())
	}
}

/*
@GetDefis retourne l'ensemble des défis de la table Defis
*/
func GetDefis() []modele.Defi {
	var (
		debutString string
		finString   string
		defi        modele.Defi
		time        time.Time = time.Now()
	)
	defis := make([]modele.Defi, 0)
	row, err := db.Query("SELECT * FROM Defis")
	if err != nil {
		logs.WriteLog("DAO.GetDefis", err.Error())
	}
	for row.Next() {
		err = row.Scan(&defi.Num, &debutString, &finString, &defi.JeuDeTest, &defi.Correcteur)
		if err != nil && defi.Correcteur != "" {
			logs.WriteLog("DAO GetDefis", err.Error())
		}
		if err = time.UnmarshalText([]byte(debutString)); err != nil {
			fmt.Println(err.Error())
		}
		defi.DateDebut = time
		if err = time.UnmarshalText([]byte(finString)); err != nil {
			fmt.Println(err.Error())
		}
		defi.DateFin = time
		defis = append(defis, defi)
		defi = modele.Defi{}
	}
	row.Close()
	if len(defis) == 0 {
		return nil
	}
	return defis
}

/*
@GetDefiActuel
@return le premier defi qui a sa date de début antérieur et sa date de fin postérieur à la date d'aujourd'hui
		defi "vide" s'il ne trouve pas de défi avec les conditions
*/
func GetDefiActuel() modele.Defi {
	defis := GetDefis()

	defiActuel := modele.Defi{
		Num:        0,
		DateDebut:  time.Time{},
		DateFin:    time.Time{},
		JeuDeTest:  false,
		Correcteur: "",
	}
	for _, d := range defis {
		if time.Now().Sub(d.DateDebut) > 0 && time.Now().Sub(d.DateFin) < 0 {
			defiActuel = d
		}
	}
	return defiActuel
}

/*
@GetDefi retourne un défi
@num numéro du défi qu'on cherche
*/
func GetDefi(num int) modele.Defi {
	defis := GetDefis()
	for _, d := range defis {
		if d.Num == num {
			return d
		}
	}
	return modele.Defi{}
}

/**
@DeleteLastDefi delete le dernier defi de la table Defi et delete tous les résultats correspondant à ce défi
@num numéro du défi supprimé
*/
func DeleteLastDefi(num int) {
	stmt, err := db.Prepare("DELETE FROM Defis WHERE numero = ?")
	_, err = stmt.Exec(num)
	if err != nil {
		logs.WriteLog("DAO.DeleteDefi", err.Error())
	}
	stmt.Close()
	stmt, err = db.Prepare("UPDATE SQLITE_SEQUENCE SET SEQ= ? WHERE NAME='Defis'")
	if err != nil {
		logs.WriteLog("Reset SQLITE SEQ", err.Error())
	}
	_, err = stmt.Exec(num - 1)
	if err != nil {
		logs.WriteLog("Reset SQLITE SEQ", err.Error())
	}
	stmt.Close()
	stmt, err = db.Prepare("DELETE FROM Resultat WHERE defi = ?")
	_, err = stmt.Exec(num)
	if err != nil {
		logs.WriteLog("DAO.DeleteDefi", err.Error())
	}
	stmt.Close()

}

/**
@AddJeuDeTest change le booléen "jeuDeTest" d'un défi pour le mettre à vrai
@num numéro du défi dont on veut spécifier qu'il possède désormais un jeu de test
*/
func AddJeuDeTest(num int) {
	stmt, err := db.Prepare("UPDATE Defis SET jeuDeTest = ? where numero = ?")
	if err != nil {
		logs.WriteLog("DAO.AddJeuDeTest n°"+strconv.Itoa(num), err.Error())
	}
	if _, err := stmt.Exec(true, num); err != nil {
		logs.WriteLog("DAO.AddJeuDeTest defi "+strconv.Itoa(num), err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("DAO.AddJeuDeTest", err.Error())
	}
}

/**
@ResetEtatDefi
@num numéro du défi
*/
func ResetEtatDefi(num int) {
	m.Lock()
	stmt, _ := db.Prepare("UPDATE Resultat SET etat = -1 WHERE defi = ?")
	if _, err := stmt.Exec(num); err != nil {
		logs.WriteLog("DAO.ResetEtatDefi "+strconv.Itoa(num), err.Error())
	}
	stmt.Close()
	m.Unlock()
}

//selectionne quel étudiant sera correcteur en fonction de si il a réussi et si il a déjà été correcteur
/**
@GenerateCorrecteur selectionne un étudiant parmis les étudiants ayant réussi le défi et l'attribue comme correcteur de ce défi
		La colonne Correcteur de l'étudiant devient True
		La colonne Correcteur du défi prend le login du correcteur
@numDefi numéro du défi auquel on génère un correcteur
*/
func GenerateCorrecteur(numDefi int) {
	m.Lock()
	var t = make([]modele.Etudiant, 0)
	var etu modele.Etudiant
	row, err := db.Query("Select e.* FROM Resultat r, Etudiant e WHERE r.Defi = $1 AND r.Etat = 1 AND e.Correcteur= 0 AND r.Login =e.Login", numDefi)
	if err != nil {
		logs.WriteLog("DAO GenerateCorrecteur", err.Error())
	} else {
		for row.Next() {
			err = row.Scan(&etu.Login, &etu.Password, &etu.Nom, &etu.Prenom, &etu.Correcteur, &etu.ResDefiActuel)
			if err != nil {
				logs.WriteLog("DAO GenerateCorrecteur", err.Error())
			}
			t = append(t, etu)
			etu = modele.Etudiant{}
		}
	}
	row.Close()

	if len(t) == 0 {
		m.Unlock()
		logs.WriteLog("GetCorrecteur", "Aucun correcteur n'a pu être choisi car aucun étudiant n'a passé le défi")
		return
	}
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := len(t) - 1
	n := max - min + 1
	aleatoire := rand.Intn(n) + min
	correcteur := t[aleatoire]
	sqlStatement := "UPDATE Etudiant  SET correcteur = 1 WHERE login = $1 "
	db.Exec(sqlStatement, correcteur.Login)
	sqlStatement = "UPDATE Defis SET correcteur = $1 WHERE numero = $2"
	db.Exec(sqlStatement, correcteur.Login, numDefi)
	fmt.Println("correcteur généré : ", correcteur)
	m.Unlock()
}

/**
@GetCorrecteur récupère le correcteur du défi
@num numéro du défi
*/
func GetCorrecteur(num int) modele.Etudiant {
	var etu modele.Etudiant
	row := db.QueryRow("SELECT e.* FROM Etudiant e, Defis d WHERE d.numero = $1 AND e.login = d.correcteur", num)
	if err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Correcteur, &etu.ResDefiActuel); err != nil {
		logs.WriteLog("DAO.GetCorrecteur", err.Error())
	}
	return etu
}

/**
@GetResultatsByEtu récupère tous les résultats de tous les défis pour un unique étudiant
@login le login de l'étudiant
*/
func GetResultatsByEtu(login string) []modele.Resultat {
	var res modele.Resultat
	resT := make([]modele.Resultat, 0)
	row, err := db.Query("SELECT * FROM Resultat WHERE login = ? ORDER BY defi ASC", login)
	if err != nil {
		logs.WriteLog("DAO GetResultatsByEtu", err.Error())
	}
	for row.Next() {
		err = row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative)
		if err != nil {
			logs.WriteLog("DAO GetResultatsByEtu", err.Error())
		}
		resT = append(resT, res)
	}
	row.Close()
	return resT
}

/**
@GetResultatsByEtat retourne le résultat de tous les étudiants ayant participé au défi numDefi et ayant eu l'état spécifié
@numDefi numéro du défi
@etat état attendu
*/
func GetResultatsByEtat(numDefi int, etat int) []modele.Resultat {
	var res modele.Resultat
	resT := make([]modele.Resultat, 0)
	row, err := db.Query("SELECT * FROM Resultat WHERE defi = ? AND etat = ?", numDefi, etat)
	if err != nil {
		logs.WriteLog("DAO GetResultatsByEtu", err.Error())
	}
	for row.Next() {
		err = row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative)
		if err != nil {
			logs.WriteLog("DAO GetResultatsByEtu", err.Error())
		}
		resT = append(resT, res)
	}
	row.Close()
	return resT
}

/**
@GetParticipants retourne la liste des participants à un défi
@numDefi numéro du défi
*/
func GetParticipants(numDefi int) []modele.ParticipantDefi {
	var res modele.ParticipantDefi
	resT := make([]modele.ParticipantDefi, 0)

	row, err := db.Query("SELECT e.login, e.prenom, e.nom, r.defi, r.etat, r.tentative FROM Etudiant e, Resultat r WHERE e.login = r.login AND r.defi = ? ORDER BY nom", numDefi)
	if err != nil {
		logs.WriteLog("DAO.GetParticipants", err.Error())
	}
	for row.Next() {
		err = row.Scan(&res.Etudiant.Login, &res.Etudiant.Prenom, &res.Etudiant.Nom, &res.Resultat.Defi, &res.Resultat.Etat, &res.Resultat.Tentative)
		//err = row.Scan(&res.Etudiant.Login, &res.Etudiant.Password, &res.Etudiant.Prenom, &res.Etudiant.Nom,
		//	&res.Etudiant.Correcteur, &res.Etudiant.ResDefiActuel,
		//	&res.Resultat.Login, &res.Resultat.Defi, &res.Resultat.Etat, &res.Resultat.Tentative)
		if err != nil {
			logs.WriteLog("DAO.GetParticipants", err.Error())
		}
		resT = append(resT, res)
		res = modele.ParticipantDefi{}
	}
	row.Close()
	return resT
}

/**
@GetEtudiantsMail récupère les mails des étudiants
*/
func GetEtudiantsMail() []modele.EtudiantMail {
	var res modele.EtudiantMail
	resT := make([]modele.EtudiantMail, 0)

	row, err := db.Query("SELECT login, prenom, nom FROM Etudiant;")
	if err != nil {
		logs.WriteLog("DAO.GetEtudiantsMail", err.Error())
	} else if row != nil {
		for row.Next() {
			err = row.Scan(&res.Login, &res.Prenom, &res.Nom)
			if err != nil {
				panic(err)
			}
			resT = append(resT, res)
			res = modele.EtudiantMail{}
		}
	}

	for i, etu := range resT {
		etu.Defis = GetResultatsByEtu(etu.Login)
		resT[i] = etu
	}
	return resT
}
