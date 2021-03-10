package BDD

import (
	"database/sql"
	"fmt"
	"github.com/aodin/date"
	_ "github.com/aodin/date"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// Structures a réutiliser un peu partout
type Etudiant struct {
	Login      string
	Password   string
	Prenom     string
	Nom        string
	Mail       string
	Correcteur bool
}

type Admin struct {
	Login    string
	Password string
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
	JeuDeTest  bool
	Correcteur string
}

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
		"mail TEXT NOT NULL," +
		"correcteur BOOLEAN NOT NULL" +
		");")
	if err != nil {
		fmt.Println("prblm table Etudiant" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Defis (" +
		"numero INTEGER PRIMARY KEY AUTOINCREMENT," +
		"date_debut TEXT NOT NULL," +
		"date_fin TEXT NOT NULL," +
		"jeu_de_test BOOL NOT NULL," +
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

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Administrateur (" +
		"login TEXT NOT NULL PRIMARY KEY ," +
		"password TEXT NOT NULL" +
		")")
	if err != nil {
		fmt.Println("Erreur dans la table Administrateur" + err.Error())
	}
	stmt.Exec()

	stmt.Close()

	admin := Admin{
		Login:    "admin",
		Password: "admin",
	}
	RegisterAdmin(admin)
}

/**
Enregistre un étudiant dans la table Etudiant
*/
func Register(etu Etudiant) bool {
	m.Lock()
	stmt, err := db.Prepare("INSERT INTO Etudiant values(?,?,?,?,?,?)")
	if err != nil {
		logs.WriteLog("BDD register étudiant : ", err.Error())
	}
	_, err = stmt.Exec(etu.Login, etu.Password, etu.Prenom, etu.Nom, etu.Mail, false)
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
func RegisterAdmin(admin Admin) bool {
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
func GetEtudiant(id string) Etudiant {
	var etu Etudiant
	row := db.QueryRow("SELECT * FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Mail, &etu.Correcteur)
	if err != nil {
		logs.WriteLog("BDD.GetEtudiant", err.Error())
	}
	return etu
}

/**
récupère les informations d'un admin
*/
func GetAdmin(id string) Admin {
	var admin Admin
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
func SaveResultat(lelogin string, lenum_defi int, letat int, admin bool) {
	m.Lock()
	var res ResBDD
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", lelogin, lenum_defi)

	if err := row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative); err != nil {
		stmt, _ := db.Prepare("INSERT INTO Resultat values(?,?,?,?)")
		_, err = stmt.Exec(lelogin, lenum_defi, letat, 1)
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
				_, err = stmt.Exec(res.Etat, res.Tentative, res.Login, res.Defi)
				if err != nil {
					logs.WriteLog("BDD SaveResultats", err.Error())
				}

			} else {
				_, err = stmt.Exec(res.Etat, res.Tentative+1, res.Login, res.Defi)
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

/**
Récupère la liste des étudiants de la table Etudiant
*/
func GetEtudiants() []Etudiant {
	var etu Etudiant
	etudiants := make([]Etudiant, 0)
	row, err := db.Query("SELECT * FROM Etudiant")
	defer row.Close()
	if err != nil {
		logs.WriteLog("BDD GetEtudiants", err.Error())
	}
	for row.Next() {
		err = row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Mail, &etu.Correcteur)
		if err != nil {
			logs.WriteLog("BDD GetEtudiants", err.Error())
		}
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
func AddDefi(dateD date.Date, dateF date.Date) {
	m.Lock()
	stmt, err := db.Prepare("INSERT INTO Defis(date_debut,date_fin, jeu_de_test) values(?,?,?)")
	if err != nil {
		logs.WriteLog("BDD AddeDefi", err.Error())
	} else {
		_, err = stmt.Exec(dateD.String(), dateF.String(), false)
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
func ModifyDefi(num int, dateD date.Date, dateF date.Date) {
	stmt, err := db.Prepare("UPDATE Defis SET date_debut = ?, date_fin = ? where numero = ?")
	if err != nil {
		logs.WriteLog("BDD modify defi", err.Error())
	}
	if _, err := stmt.Exec(dateD.String(), dateF.String(), num); err != nil {
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
		err = row.Scan(&defi.Num, &debutString, &finString, &defi.JeuDeTest, &defi.Correcteur)
		if err != nil && defi.Correcteur != "" {
			logs.WriteLog("BDD GetDefis", err.Error())
		}
		defi.Date_debut, _ = date.Parse(debutString)
		defi.Date_fin, _ = date.Parse(finString)
		defis = append(defis, defi)
		defi = Defi{}
	}
	if len(defis) == 0 {
		return nil
	}
	return defis
}

/*
Récupère le défi actuel
*/
func GetDefiActuel() Defi {
	defis := GetDefis()

	defiActuel := Defi{
		Num:        0,
		Date_debut: date.Date{},
		Date_fin:   date.Date{},
		Correcteur: "",
	}
	for _, d := range defis {
		if date.Today().Within(date.NewRange(d.Date_debut, d.Date_fin)) {
			defiActuel = d
		}
	}
	return defiActuel
}

/*
Récupère un défi précis, avec son numéro passé en paramètre
*/

func GetDefi(num int) Defi {
	defis := GetDefis()
	for _, d := range defis {
		if d.Num == num {
			return d
		}
	}
	return Defi{}
}

func AddJeuDeTest(num int) {
	stmt, err := db.Prepare("UPDATE Defis SET jeu_de_test = ? where numero = ?")
	if err != nil {
		logs.WriteLog("BDD ajout d'un jeu de test au défi n°"+strconv.Itoa(num), err.Error())
	}
	if _, err := stmt.Exec(true, num); err != nil {
		logs.WriteLog("BDD.AddJeuDeTest", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		logs.WriteLog("BDD.AddJeuDeTest", err.Error())
	}
}

//selectionne quel étudiant sera correcteur en fonction de si il a réussi et si il a déjà été correcteur
func GenerateCorrecteur(num_defi int) {
	m.Lock()
	var t = make([]Etudiant, 0)
	var etu Etudiant
	row, err := db.Query("Select e.* FROM Resultat r, Etudiant e WHERE r.Defi = $1 AND r.Etat = 1 AND e.Correcteur= 0 AND r.Login =e.Login", num_defi)
	defer row.Close()
	if err != nil {
		logs.WriteLog("BDD GenerateCorrecteur", err.Error())
	} else {
		for row.Next() {
			err = row.Scan(&etu.Login, &etu.Password, &etu.Nom, &etu.Prenom, &etu.Mail, &etu.Correcteur)
			if err != nil {
				logs.WriteLog("BDD GenerateCorrecteur", err.Error())
			}
			t = append(t, etu)
			etu = Etudiant{}
		}
	}
	fmt.Println(t)
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := len(t) - 1
	n := max - min + 1
	aleatoire := rand.Intn(n) + min
	correcteur := t[aleatoire]
	sqlStatement := "UPDATE Etudiant  SET correcteur = 1 WHERE login = $1 "
	db.Exec(sqlStatement, correcteur.Login)
	sqlStatement = "UPDATE Defis SET correcteur = $1 WHERE numero = $2"
	db.Exec(sqlStatement, correcteur.Login, num_defi)
	fmt.Println("correcteur généré : ", correcteur)
	m.Unlock()
}

func GetCorrecteur(num int) Etudiant {
	var etu Etudiant
	row := db.QueryRow("SELECT e.* FROM Etudiant e, Defis d WHERE d.numero = $1 AND e.login = d.correcteur", num)
	if err := row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Mail, &etu.Correcteur); err != nil {
		logs.WriteLog("BDD.GetCorrecteur", err.Error())
	}
	return etu
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
		logs.WriteLog("BDD GetAllResultat", err.Error())
	}
	for row.Next() {
		err = row.Scan(&res.Login, &res.Defi, &res.Etat, &res.Tentative)
		if err != nil {
			logs.WriteLog("BDD GetAllResultat", err.Error())
		}
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
		err = row.Scan(&res.Etudiant.Login, &res.Etudiant.Password, &res.Etudiant.Prenom, &res.Etudiant.Nom, &res.Etudiant.Mail, &res.Etudiant.Correcteur, &res.Resultat.Login, &res.Resultat.Defi,
			&res.Resultat.Etat, &res.Resultat.Tentative)
		if err != nil {
			logs.WriteLog("Bdd GetParticipants", err.Error())
		}
		resT = append(resT, res)
		res = ParticipantDefi{}
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
			res = EtudiantMail{}
		}
	}

	for i, etu := range resT {
		etu.Defis = GetAllResultat(etu.Login)
		resT[i] = etu
	}
	return resT
}
