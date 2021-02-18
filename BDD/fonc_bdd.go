package BDD

import (
	"database/sql"
	"fmt"
	"github.com/aodin/date"
	_ "github.com/aodin/date"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

// Structure a réutiliser un peu partout
type Etudiant struct {
	Login      string
	Password   string
	Prenom     string
	Nom        string
	Mail       string
	Correcteur bool
}

type EtudiantMail struct {
	Login  string
	Prenom string
	Nom    string
	Mail   string
	Defis  []ResBDD
}

type ResBDD struct {
	Login     string
	Defi      int
	Etat      int
	Tentative int
}
type ParticipantDefi struct {
	Etudiant Etudiant
	Resultat ResBDD
}

type Defi struct {
	Num        int
	Date_debut date.Date
	Date_fin   date.Date
}

var db, _ = sql.Open("sqlite3", "./BDD/database.db")

/**
Fonction qui initialise les tables vides
*/
func InitBDD() {

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS Etudiant (" +
		"login TEXT PRIMARY KEY, " +
		"password TEXT NOT NULL, " +
		"prenom TEXT NOT NULL," +
		"nom TEXT NOT NULL," +
		"mail TEXT NOT NULL," +
		"correcteur BOOLEAN NOT NULL" +
		");")
	if err != nil {
		fmt.Println("prblm table Etudiant" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Defis (" +
		"numero INTEGER PRIMARY KEY," +
		"date_debut TEXT NOT NULL," +
		"date_fin TEXT NOT NULL" +
		")")
	if err != nil {
		fmt.Println("Erreur création table Defis " + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Resultat (" +
		"login TEXT NOT NULL," +
		"defi INTEGER NOT NULL," +
		"etat INTEGER NOT NULL," + // 2 états : 1 (réussi), 0 (non réussi), -1 (non testé)
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

	stmt.Close()
}

/**
Enregistre un étudiant dans la table Etudiant
*/
func Register(etu Etudiant) bool {
	stmt, _ := db.Prepare("INSERT INTO Etudiant values(?,?,?,?,?,?)")

	_, err := stmt.Exec(etu.Login, etu.Password, etu.Prenom, etu.Nom, etu.Mail, false)
	if err != nil {
		logs.WriteLog("BDD.Register", err.Error())
		return false
	}
	logs.WriteLog("Register", etu.Login+" est enregistré")
	stmt.Close()
	return true
}

/**
vérifie que le couple login,password existe dans la table Etudiant
*/
func LoginCorrect(id string, password string) bool {
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
	return errCompare == nil                                                              // si nil -> ça match, sinon non

	/* Ancient système
	stmt := "SELECT * FROM Etudiant WHERE login = ? AND password = ?"
	row, _ := db.Query(stmt, id, password)
	if row.Next() {
		row.Close()
		return true
	}
	row.Close()
	return false*/
}

/**
récupère les informations personnelles d'un étudiant
*/
func GetEtudiant(id string) Etudiant {
	var etu Etudiant
	row := db.QueryRow("SELECT * FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Mail, &etu.Correcteur)

	if err != nil {
		logs.WriteLog("BDD.GetEtudiant", err.Error())
	}
	return etu
}

// testé
func GetNameByToken(token string) string {
	var login string
	row := db.QueryRow("SELECT * FROM token WHERE token = $1", token)
	err := row.Scan(&login, &token)
	if err != nil {
		logs.WriteLog("BDD.GetNameByToken", err.Error())
	}
	return login
}

// testé
func InsertToken(login string, token string) {

	stmt, _ := db.Prepare("DELETE FROM Token where login = ?")
	stmt.Exec(login)
	stmt, _ = db.Prepare("INSERT INTO Token values(?,?)")
	_, err := stmt.Exec(login, token)
	if err != nil {
		logs.WriteLog("BDD.InsertToken", err.Error())
	}
	stmt.Close()
}

// testé
func DeleteToken(login string) {
	stmt, _ := db.Prepare("DELETE FROM token WHERE login = ?")
	_, err := stmt.Exec(login)
	if err != nil {
		logs.WriteLog("BDD.DeleteToken", err.Error())
	}
	stmt.Close()
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
func ResetToken() {
	stmt, _ := db.Prepare("TRUNCATE TABLE token;")
	if _, err := stmt.Exec(); err != nil {
		fmt.Printf("erreur clear de la table token")
	}
	stmt.Close()
}

/**
admin == true : fonction lancé par l'admin pour modifier les valeurs
admin == false : fonction lancé par un étudiant lors d'une nouvelle tentative de test
(si c'est false, tentative++)
*/
func SaveResultat(lelogin string, lenum_defi int, letat int, admin bool) {
	var res ResBDD
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", lelogin, lenum_defi)

	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		stmt, _ := db.Prepare("INSERT INTO Resultat values(?,?,?,?)")
		_, err = stmt.Exec(lelogin, lenum_defi, letat, 1)
		stmt.Close()
	} else {
		stmt, _ := db.Prepare("UPDATE Resultat SET etat = ?, tentative = ? WHERE login = ? AND defi = ?")
		if admin {
			stmt.Exec(res.Etat, res.Tentative, res.Login, res.Defi)
		} else {
			stmt.Exec(res.Etat, res.Tentative+1, res.Login, res.Defi)
		}
		stmt.Close()
	}

}

/**
Récupère la liste des étudiants de la table Etudiant
*/
func GetEtudiants() []Etudiant {
	var etu Etudiant
	etudiants := make([]Etudiant, 0)
	row, err := db.Query("SELECT * FROM Etudiant")
	defer row.Close()
	if err != nil {
		logs.WriteLog("BDD.GetEtudiants", err.Error())
	}
	for row.Next() {
		row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Mail, &etu.Correcteur)
		etudiants = append(etudiants, etu)
	}
	return etudiants
}

/**
Récupère le résultat d'un étudiant pour un défi spécifique
*/
func GetResult(login string, defi int) ResBDD {
	var res ResBDD
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", login, defi)
	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		logs.WriteLog("BDD.GetResult", err.Error())
	}
	return res
}

/**
Ajoute un défi à la table Defis
*/
func AddDefi(num int, dateD date.Date, dateF date.Date) {
	stmt, err := db.Prepare("INSERT INTO Defis values(?,?,?)")
	_, err = stmt.Exec(num, dateD.String(), dateF.String())
	if err != nil {
		logs.WriteLog("BDD.AddDefi", err.Error())
	}
	stmt.Close()
}

/**
Modifie le défi de numéro num
*/
func ModifyDefi(num int, dateD date.Date, dateF date.Date) {
	stmt, _ := db.Prepare("UPDATE Defis SET date_debut = ?, date_fin = ? where numero = ?")
	if _, err := stmt.Exec(dateD.String(), dateF.String(), num); err != nil {
		logs.WriteLog("BDD.ModifyDefi", err.Error())
	}
	stmt.Close()
}

func GetDefis() []Defi {
	var (
		debutString string
		finString   string
		defi        Defi
	)
	defis := make([]Defi, 0)
	row, err := db.Query("SELECT * FROM Defis")
	defer row.Close()
	if err != nil {
		logs.WriteLog("BDD.GetDefis", err.Error())
	}
	for row.Next() {
		row.Scan(&defi.Num, &debutString, &finString)
		defi.Date_debut, _ = date.Parse(debutString)
		defi.Date_fin, _ = date.Parse(finString)
		defis = append(defis, defi)
	}
	return defis
}

func GetDefiActuel() Defi {
	defis := GetDefis()

	defiActuel := Defi{
		Num:        -1,
		Date_debut: date.Date{},
		Date_fin:   date.Date{},
	}
	for _, d := range defis {
		if date.Today().Within(date.NewRange(d.Date_debut, d.Date_fin)) {
			defiActuel = d
		}
	}
	return defiActuel
}

//selectionne quel étudiant sera correcteur en fonction de si il a réussi et si il a déjà été correcteur
func GetEtudiantCorrecteur(num_defi int) string {
	var t = make([]string, 0)
	var res string
	var aleatoire int
	var logfinal string
	row, err := db.Query("Select Login FROM Resultat r, Etudiant e WHERE r.Defi =", num_defi, " AND r.Etat = 1 AND e.Correcteur= 0 AND r.Login =e.Login")
	defer row.Close()
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		for row.Next() {
			row.Scan(&res)
			t = append(t, res)
		}
		aleatoire = rand.Intn(len(t))
	}
	logfinal = t[aleatoire]
	return logfinal
}

/**
Récupère tous les résultats d'un étudiant à tous les défis auquel il a participé
*/
func GetAllResultat(login string) []ResBDD {
	var res ResBDD
	resT := make([]ResBDD, 0)
	row, err := db.Query("SELECT * FROM Resultat WHERE login = ? ORDER BY defi ASC", login)
	defer row.Close()
	if err != nil {
		logs.WriteLog("BDD.GetAllResultat", err.Error())
	}
	for row.Next() {
		row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative)
		resT = append(resT, res)
	}

	return resT
}

/**
Récupère tous les résultats de tous les étudiants pour un défi spécifique
*/
func GetParticipant(num_defi int) []ParticipantDefi {
	var res ParticipantDefi
	resT := make([]ParticipantDefi, 0)

	row, err := db.Query("SELECT * FROM Etudiant e, Resultat r WHERE e.login = r.login AND r.defi = ? ORDER BY nom", num_defi)
	defer row.Close()
	if err != nil {
		logs.WriteLog("BDD.GetParticipant", err.Error())
	}
	for row.Next() {
		row.Scan(&res.Etudiant.Login, &res.Etudiant.Password, &res.Etudiant.Prenom, &res.Etudiant.Nom, &res.Etudiant.Mail, &res.Etudiant.Correcteur, &res.Resultat.Login, &res.Resultat.Defi,
			&res.Resultat.Etat, &res.Resultat.Tentative)
		resT = append(resT, res)
	}
	return resT
}

func GetEtudiantsMail() []EtudiantMail {
	var res EtudiantMail
	resT := make([]EtudiantMail, 0)

	row, err := db.Query("SELECT  login, prenom, nom, mail FROM Etudiant;")
	if err != nil {
		logs.WriteLog("BDD.GetEtudiantsMail", err.Error())
	} else if row != nil {
		for row.Next() {
			err = row.Scan(&res.Login, &res.Prenom, &res.Nom, &res.Mail)
			if err != nil {
				panic(err)
			}
			resT = append(resT, res)
		}
	}

	for i, etu := range resT {
		etu.Defis = GetAllResultat(etu.Login)
		resT[i] = etu
	}
	return resT
}
