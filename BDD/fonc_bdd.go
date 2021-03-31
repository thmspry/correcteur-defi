package BDD

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var db, _ = sql.Open("sqlite3", "./BDD/database.db")
var m sync.Mutex

/**
Fonction qui initialise les tables vides
*/
func InitBDD() {

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

	admin := config.Admin{
		Login:    "admin",
		Password: "admin",
	}
	RegisterAdmin(admin)
}

/**
Enregistre un étudiant dans la table Etudiant
*/
func Register(etu config.Etudiant) bool {
	m.Lock()
	stmt, err := db.Prepare("INSERT INTO Etudiant(login,password,prenom,nom,correcteur) values(?,?,?,?,?)")
	if err != nil {
		logs.WriteLog("BDD register étudiant : ", err.Error())
	}
	_, err = stmt.Exec(etu.Login, etu.Password, etu.Prenom, etu.Nom, false)
	if err != nil {
		logs.WriteLog("BDD register étudiant : ", err.Error())
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
func RegisterAdmin(admin config.Admin) bool {
	m.Lock()
	stmt, err := db.Prepare("INSERT INTO Administrateur values(?,?)")
	if err != nil {
		logs.WriteLog("BDD register admin : ", err.Error())
	}
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(admin.Password), 14)
	_, err = stmt.Exec(admin.Login, passwordHashed)
	if err != nil {
		logs.WriteLog("BDD register admin : ", err.Error())
		m.Unlock()
		return false
	}
	fmt.Println("l'admin de login : " + admin.Login + " a été enregistré dans la bdd\n")
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("BDD register admin : ", err.Error())
	}
	m.Unlock()
	return true
}

/**

 */
func RegisterAdminString(login string, password string) bool {
	admin := config.Admin{
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
		logs.WriteLog("BDD.LoginCorrect", err.Error())
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
func GetEtudiant(id string) config.Etudiant {
	var etu config.Etudiant
	row := db.QueryRow("SELECT * FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Correcteur, &etu.ResDefiActuel)
	if err != nil && etu.ResDefiActuel != nil {
		logs.WriteLog("BDD.GetEtudiant", err.Error())
	}
	return etu
}

/**
récupère les informations d'un admin
*/
func GetAdmin(id string) config.Admin {
	var admin config.Admin
	row := db.QueryRow("SELECT * FROM Administrateur WHERE login = $1", id)
	err := row.Scan(&admin.Login, &admin.Password)

	if err != nil {
		logs.WriteLog("BDD GetAdmin "+id+" : ", err.Error())
	}
	return admin
}

/**

 */
func GetNameByToken(token string) string {
	var login string
	row := db.QueryRow("SELECT * FROM token WHERE token = $1", token)
	err := row.Scan(&login, &token)
	if err != nil {
		logs.WriteLog("BDD GetNameByToken "+token+" : ", err.Error())
	}
	return login
}

// testé
func InsertToken(login string, token string) {
	m.Lock()
	stmt, _ := db.Prepare("DELETE FROM Token where login = ?")
	_, err := stmt.Exec(login)
	if err != nil {
		logs.WriteLog("BDD InsertToken "+login, err.Error())
	}
	stmt, err = db.Prepare("INSERT INTO Token values(?,?)")
	if err != nil {
		logs.WriteLog("BDD InsertToken "+login, err.Error())
	}
	_, err = stmt.Exec(login, token)
	if err != nil {
		logs.WriteLog("BDD InsertToken "+login, err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("BDD InsertToken "+login, err.Error())
	}
	m.Unlock()
}

// testé
func DeleteToken(login string) {
	m.Lock()
	stmt, err := db.Prepare("DELETE FROM token WHERE login = ?")
	if err != nil {
		logs.WriteLog("BDD.DeleteToken", err.Error())
	}
	_, err = stmt.Exec(login)
	if err != nil {
		logs.WriteLog("BDD DeleteToken", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("BDD.DeleteToken", err.Error())
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

	logs.WriteLog("BDD TokenRole", "Le login associé n'est pas dans la table administrateur ou étudiant")
	return ""
}

/**
admin == true : fonction lancé par l'admin pour modifier les valeurs
admin == false : fonction lancé par un étudiant lors d'une nouvelle tentative de test
(si c'est false, tentative++)
*/
func SaveResultat(login string, numDefi int, etat int, resultat []config.Resultat, admin bool) {
	m.Lock()

	if resultat != nil { //ajout du résultat obtenu à l'étudiant
		resJson, _ := json.Marshal(resultat)
		stmt, err := db.Prepare("UPDATE Etudiant SET resDefiActuel = ? WHERE login = ?")
		if err != nil {
			logs.WriteLog("BDD.SaveResultat", err.Error())
		}
		_, err = stmt.Exec(string(resJson), login)
		if err != nil {
			logs.WriteLog("BDD DeleteToken", err.Error())
		}
		err = stmt.Close()
	}

	var res config.ResBDD
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", login, numDefi)

	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		//si err diff nil, cela veut dire qu'il n'a pas réussi à scan car il n'y a pas de ligne dans row
		stmt, _ := db.Prepare("INSERT INTO Resultat values(?,?,?,?)")
		_, err = stmt.Exec(login, numDefi, etat, 1)
		err = stmt.Close()
		if err != nil {
			logs.WriteLog("BDD SaveResultat", err.Error())
		}
	} else {
		stmt, err := db.Prepare("UPDATE Resultat SET etat = ?, tentative = ? WHERE login = ? AND defi = ?")
		if err != nil {
			logs.WriteLog("BDD SaveResultats", err.Error())
		} else {
			if admin {
				_, err = stmt.Exec(etat, res.Tentative, res.Login, res.Defi)
				if err != nil {
					logs.WriteLog("BDD SaveResultats", err.Error())
				}

			} else {
				_, err = stmt.Exec(etat, res.Tentative+1, res.Login, res.Defi)
				if err != nil {
					logs.WriteLog("BDD SaveResultats", err.Error())
				}
			}
			err = stmt.Close()
			if err != nil {
				logs.WriteLog("BDD SaveResultats", err.Error())
			}
		}
	}
	m.Unlock()
}

func GetResultatActuel(login string) []config.Resultat {
	var (
		query string
		res   []config.Resultat
	)
	row := db.QueryRow("SELECT resDefiActuel FROM Etudiant WHERE login = $1", login)
	if err := row.Scan(&query); err != nil {
		logs.WriteLog("BDD.GetResultatActuel", err.Error())
	}
	json.Unmarshal([]byte(query), &res)
	return res
}

/**
Récupère la liste des étudiants de la table Etudiant
*/
func GetEtudiants() []config.Etudiant {
	var etu config.Etudiant
	etudiants := make([]config.Etudiant, 0)
	row, err := db.Query("SELECT * FROM Etudiant")
	if err != nil {
		logs.WriteLog("BDD GetEtudiants", err.Error())
	}
	for row.Next() {
		err = row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Correcteur, &etu.ResDefiActuel)
		if err != nil && etu.ResDefiActuel != nil {
			logs.WriteLog("BDD GetEtudiants", err.Error())
		}
		etudiants = append(etudiants, etu)
	}
	row.Close()
	return etudiants
}

/**
Récupère le résultat d'un étudiant pour un défi spécifique
*/
func GetResult(login string, defi int) config.ResBDD {
	var res config.ResBDD
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", login, defi)
	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		logs.WriteLog("BDD.GetResult", err.Error())
	}
	return res
}

/**
Ajoute un défi à la table Defis
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
		logs.WriteLog("BDD AddeDefi", err.Error())
	} else {
		_, err = stmt.Exec(string(tDeb), string(tFin), false)
		if err != nil {
			logs.WriteLog("BDD AddDefi", err.Error())
		}
		stmt.Close()
	}
	m.Unlock()
}

/**
Modifie le défi de numéro num
*/
func ModifyDefi(num int, dateD time.Time, dateF time.Time) {
	tDeb, err := dateD.MarshalText()
	tFin, err := dateF.MarshalText()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt, err := db.Prepare("UPDATE Defis SET dateDebut = ?, dateFin = ? where numero = ?")
	if err != nil {
		logs.WriteLog("BDD modify defi", err.Error())
	}
	if _, err := stmt.Exec(string(tDeb), string(tFin), num); err != nil {
		logs.WriteLog("BDD.ModifyDefi", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("BDD modify defi", err.Error())
	}
}

/*
Récupère la liste des défis
*/
func GetDefis() []config.Defi {
	var (
		debutString string
		finString   string
		defi        config.Defi
		time        time.Time = time.Now()
	)
	defis := make([]config.Defi, 0)
	row, err := db.Query("SELECT * FROM Defis")
	if err != nil {
		logs.WriteLog("BDD.GetDefis", err.Error())
	}
	for row.Next() {
		err = row.Scan(&defi.Num, &debutString, &finString, &defi.JeuDeTest, &defi.Correcteur)
		if err != nil && defi.Correcteur != "" {
			logs.WriteLog("BDD GetDefis", err.Error())
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
		defi = config.Defi{}
	}
	row.Close()
	if len(defis) == 0 {
		return nil
	}
	return defis
}

/*
Récupère le défi actuel
*/
func GetDefiActuel() config.Defi {
	defis := GetDefis()

	defiActuel := config.Defi{
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
Récupère un défi précis, avec son numéro passé en paramètre
*/

func GetDefi(num int) config.Defi {
	defis := GetDefis()
	for _, d := range defis {
		if d.Num == num {
			return d
		}
	}
	return config.Defi{}
}

func DeleteLastDefi(num int) {
	stmt, err := db.Prepare("DELETE FROM Defis WHERE numero = ?")
	_, err = stmt.Exec(num)
	if err != nil {
		logs.WriteLog("BDD.DeleteDefi", err.Error())
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
		logs.WriteLog("BDD.DeleteDefi", err.Error())
	}
	stmt.Close()

}

/**
Fonction qui va mettre jeu de test à true pour signaler qu'un jeu de test a été upload pour le défi du num donné en argument
*/
func AddJeuDeTest(num int) {
	stmt, err := db.Prepare("UPDATE Defis SET jeuDeTest = ? where numero = ?")
	if err != nil {
		logs.WriteLog("BDD.AddJeuDeTest n°"+strconv.Itoa(num), err.Error())
	}
	if _, err := stmt.Exec(true, num); err != nil {
		logs.WriteLog("BDD.AddJeuDeTest defi "+strconv.Itoa(num), err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("BDD.AddJeuDeTest", err.Error())
	}
}

/**
 * Repasse tous les résultats du défi dont le numéro est donné en argument à -1 (non testé)
 * fonction appelé lorsqu'on change le jeu de test
 */
func ResetEtatDefi(num int) {
	m.Lock()
	stmt, _ := db.Prepare("UPDATE Resultat SET etat = -1 WHERE defi = ?")
	if _, err := stmt.Exec(num); err != nil {
		logs.WriteLog("BDD.ResetEtatDefi "+strconv.Itoa(num), err.Error())
	}
	stmt.Close()
	m.Unlock()
}

//selectionne quel étudiant sera correcteur en fonction de si il a réussi et si il a déjà été correcteur
func GenerateCorrecteur(numDefi int) {
	m.Lock()
	var t = make([]config.Etudiant, 0)
	var etu config.Etudiant
	row, err := db.Query("Select e.* FROM Resultat r, Etudiant e WHERE r.Defi = $1 AND r.Etat = 1 AND e.Correcteur= 0 AND r.Login =e.Login", numDefi)
	if err != nil {
		logs.WriteLog("BDD GenerateCorrecteur", err.Error())
	} else {
		for row.Next() {
			err = row.Scan(&etu.Login, &etu.Password, &etu.Nom, &etu.Prenom, &etu.Correcteur, &etu.ResDefiActuel)
			if err != nil {
				logs.WriteLog("BDD GenerateCorrecteur", err.Error())
			}
			t = append(t, etu)
			etu = config.Etudiant{}
		}
	}
	row.Close()
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

func GetCorrecteur(num int) config.Etudiant {
	var etu config.Etudiant
	row := db.QueryRow("SELECT e.* FROM Etudiant e, Defis d WHERE d.numero = $1 AND e.login = d.correcteur", num)
	if err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Correcteur, &etu.ResDefiActuel); err != nil {
		logs.WriteLog("BDD.GetCorrecteur", err.Error())
	}
	return etu
}

/**
Récupère tous les résultats d'un étudiant à tous les défis auquel il a participé
*/
func GetAllResultat(login string) []config.ResBDD {
	var res config.ResBDD
	resT := make([]config.ResBDD, 0)
	row, err := db.Query("SELECT * FROM Resultat WHERE login = ? ORDER BY defi ASC", login)
	if err != nil {
		logs.WriteLog("BDD GetAllResultat", err.Error())
	}
	for row.Next() {
		err = row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative)
		if err != nil {
			logs.WriteLog("BDD GetAllResultat", err.Error())
		}
		resT = append(resT, res)
	}
	row.Close()
	return resT
}

/**
Récupère tous les résultats de tous les étudiants pour un défi spécifique
*/
func GetParticipant(numDefi int) []config.ParticipantDefi {
	var res config.ParticipantDefi
	resT := make([]config.ParticipantDefi, 0)

	row, err := db.Query("SELECT * FROM Etudiant e, Resultat r WHERE e.login = r.login AND r.defi = ? ORDER BY nom", numDefi)
	if err != nil {
		logs.WriteLog("BDD.GetParticipant", err.Error())
	}
	for row.Next() {
		err = row.Scan(&res.Etudiant.Login, &res.Etudiant.Password, &res.Etudiant.Prenom, &res.Etudiant.Nom,
			&res.Etudiant.Correcteur, &res.Etudiant.ResDefiActuel,
			&res.Resultat.Login, &res.Resultat.Defi, &res.Resultat.Etat, &res.Resultat.Tentative)
		if err != nil {
			logs.WriteLog("Bdd GetParticipants", err.Error())
		}
		resT = append(resT, res)
		res = config.ParticipantDefi{}
	}
	row.Close()
	return resT
}

func GetEtudiantsMail() []config.EtudiantMail {
	var res config.EtudiantMail
	resT := make([]config.EtudiantMail, 0)

	row, err := db.Query("SELECT  login, prenom, nom FROM Etudiant;")
	if err != nil {
		logs.WriteLog("BDD.GetEtudiantsMail", err.Error())
	} else if row != nil {
		for row.Next() {
			err = row.Scan(&res.Login, &res.Prenom, &res.Nom)
			if err != nil {
				panic(err)
			}
			resT = append(resT, res)
			res = config.EtudiantMail{}
		}
	}

	for i, etu := range resT {
		etu.Defis = GetAllResultat(etu.Login)
		resT[i] = etu
	}
	return resT
}
