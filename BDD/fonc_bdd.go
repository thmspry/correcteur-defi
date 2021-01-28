package BDD

import (
	"database/sql"
	"fmt"
	"github.com/aodin/date"
	_ "github.com/aodin/date"
	"time"
)

// Structure a réutiliser un peu partout
type Etudiant struct {
	Login    string
	Password string
	Prenom   string
	Nom      string
	Mail     string
}

type ResBDD struct {
	Login     string
	Defi      int
	Etat      int
	Tentative int
}
type ResultatCSV struct {
	Etudiant Etudiant
	Resultat ResBDD
}

type Defi struct {
	Num        int
	Date_debut date.Date
	Date_fin   date.Date
}

var db, _ = sql.Open("sqlite3", "./BDD/projS3.db")

/**
Fonction qui initialise les tables vides
*/
func InitBDD() {

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS Etudiant (" +
		"login TEXT PRIMARY KEY, " +
		"password TEXT NOT NULL, " +
		"prenom TEXT NOT NULL," +
		"nom TEXT NOT NULL," +
		"mail TEXT NOT NULL" +
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
	stmt, err := db.Prepare("INSERT INTO Etudiant values(?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
	}

	_, err = stmt.Exec(etu.Login, etu.Password, etu.Prenom, etu.Nom, etu.Mail)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("l'étudiant de login : " + string(etu.Login) + " a été enregistré dans la bdd\n")
	stmt.Close()
	return true
}

/**
vérifie que le couple login,password existe dans la table Etudiant
*/
func LoginCorrect(id string, password string) bool {
	stmt := "SELECT * FROM Etudiant WHERE login = ? AND password = ?"
	row, _ := db.Query(stmt, id, password)
	if row.Next() {
		row.Close()
		return true
	}
	row.Close()
	return false
}

/**
récupère les informations personnelles d'un étudiant
*/
func GetEtudiant(id string) Etudiant {
	var (
		login    string
		password string
		prenom   string
		nom      string
		mail     string
	)
	row := db.QueryRow("SELECT * FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&login, &password, &prenom, &nom, &mail)

	if err != nil {
		fmt.Printf("problme row scan \n", err)
	}
	etu := Etudiant{
		Login:    login,
		Password: password,
		Prenom:   prenom,
		Nom:      nom,
		Mail:     mail,
	}
	return etu
}

// testé
func GetNameByToken(token string) string {
	var login string
	row := db.QueryRow("SELECT * FROM token WHERE token = $1", token)
	err := row.Scan(&login, &token)
	if err != nil {
		fmt.Printf("problme GetNameByToken \n", err)
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
		fmt.Println("probleme InsertToken", err)
	} else {
		fmt.Println("Ajout token : ", token, " pour ", login, " à ", time.Now(), "\n")
	}
	stmt.Close()
}

// testé
func DeleteToken(login string) {
	stmt, _ := db.Prepare("DELETE FROM token WHERE login = ?")
	_, err := stmt.Exec(login)
	if err != nil {
		fmt.Println("probleme DeleteToken", err)
	} else {
		fmt.Println("delete token du login, ", login, " à ", time.Now(), "\n")
	}
	stmt.Close()
}

func TokenExiste(token string) bool {
	row := db.QueryRow("SELECT * FROM token WHERE token = ?")
	if row.Err() != nil {
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
*/
func SaveResultat(lelogin string, lenum_defi int, letat int, admin bool) {

	var (
		login     string
		defi      int
		etat      int
		tentative int
	)
	row := db.QueryRow("SELECT * FROM Resultat WHERE login = $1 AND defi = $2", lelogin, lenum_defi)

	if err := row.Scan(&login, &defi, &etat, &tentative); err != nil {
		stmt, _ := db.Prepare("INSERT INTO Resultat values(?,?,?,?)")
		_, err = stmt.Exec(lelogin, lenum_defi, letat, 1)
		stmt.Close()
	} else {
		stmt, _ := db.Prepare("UPDATE Resultat SET etat = ?, tentative = ? WHERE login = ? AND defi = ?")
		if admin {
			stmt.Exec(letat, tentative, login, defi)
		} else {
			stmt.Exec(letat, tentative+1, login, defi)
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
		fmt.Printf(err.Error())
	}
	for row.Next() {
		row.Scan(&etu.Login, &etu.Password, &etu.Prenom, &etu.Nom, &etu.Mail)
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
		fmt.Println("erreur GetResult")
	}
	return res
}

/**
Récupère le dernier défi enregistrer dans la table Defis
*/
func GetDefiActuel() Defi {
	var (
		num   int
		debut string
		fin   string
	)
	row := db.QueryRow("SELECT * FROM Defis ORDER BY numero DESC")
	err := row.Scan(&num, &debut, &fin)
	if err != nil {
		return Defi{
			Num:        -1,
			Date_debut: date.Date{},
			Date_fin:   date.Date{},
		}
	}
	d := Defi{
		Num:        num,
		Date_debut: date.Date{},
		Date_fin:   date.Date{},
	}
	d.Date_debut, _ = date.Parse(debut)
	d.Date_fin, _ = date.Parse(fin)
	//Pas besoin de check si le date.Parse retourne une erreur car le string date enregistré dans la BDD est forcément correcte
	//étant donné qu'on vérifie qu'il soit correcte avant de l'enregistrer
	return d
}

/**
Ajoute un défi à la table Defis
*/
func AddDefi(num int, dateD date.Date, dateF date.Date) {
	stmt, err := db.Prepare("INSERT INTO Defis values(?,?,?)")
	_, err = stmt.Exec(num, dateD.String(), dateF.String())
	if err != nil {
		fmt.Println("erreur add défi " + err.Error())
	}
	stmt.Close()
}

/**
Modifie le défi de numéro num
*/
func ModifyDefi(num int, dateD date.Date, dateF date.Date) {
	stmt, err := db.Prepare("UPDATE Defis SET date_debut = ?, date_fin = ? where numero = ?")
	if err != nil {
		fmt.Println(err)
	}
	if _, err := stmt.Exec(dateD.String(), dateF.String(), num); err != nil {
		fmt.Println("erreur modify defi " + err.Error())
	}
	stmt.Close()
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
		fmt.Printf(err.Error())
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
func GetAllResult(num_defi int) []ResultatCSV {
	var res ResultatCSV
	resT := make([]ResultatCSV, 0)

	row, err := db.Query("SELECT * FROM Etudiant e, Resultat r WHERE e.login = r.login AND r.defi = ? ORDER BY nom", num_defi)
	defer row.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	for row.Next() {
		row.Scan(&res.Etudiant.Login, &res.Etudiant.Password, &res.Etudiant.Prenom, &res.Etudiant.Nom, &res.Etudiant.Mail, &res.Resultat.Login, &res.Resultat.Defi,
			&res.Resultat.Etat, &res.Resultat.Tentative)
		resT = append(resT, res)
	}

	return resT
}
