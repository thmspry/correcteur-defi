package BDD

import (
	"database/sql"
	"fmt"
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
type Defi struct {
	Num        int
	Date_debut string
	Date_fin   string
}

var db, _ = sql.Open("sqlite3", "./BDD/projS3.db")

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
		"etat INTEGER NOT NULL," + // 2 états : 1 (réussi), 0 (non réussi), (s'il n'y a pas de ligne == non tenté)
		"tentative INTEGER NOT NULL," + // Nombre de tentative au test
		"FOREIGN KEY (login) REFERENCES Etudiant(login)" +
		"FOREIGN KEY (defi) REFERENCES Defis(numero)" +
		")")
	if err != nil {
		fmt.Println("prblm table Res" + err.Error())
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

// testé
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

// testé
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

// testé
func GetInfo(id string) Etudiant {
	fmt.Println("fonc GetInfo : ")
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

	} else {
		fmt.Println("etu : ", login, password, prenom, nom, mail)
	}
	etu := Etudiant{
		Login:    login,
		Password: password,
		Prenom:   prenom,
		Nom:      nom,
		Mail:     mail,
	}

	fmt.Println("/ fonc GetInfo")
	return etu
}

// testé
func GetNameByToken(token string) string {
	fmt.Println("getNameByToken(", token, ")")
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

func ResetToken() {
	stmt, _ := db.Prepare("TRUNCATE TABLE token;")
	if _, err := stmt.Exec(); err != nil {
		fmt.Printf("erreur clear de la table token")
	}
	stmt.Close()
}

/**
admin == true : fonction lancé par l'admin pour modifier les valeurs
admin == false : fonction lancé lors d'une nouvelle tentative
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
	}
}

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

func GetLastDefi() Defi {
	var defi Defi
	row := db.QueryRow("SELECT * FROM Defis ORDER BY numero DESC")
	err := row.Scan(&defi.Num, &defi.Date_debut, &defi.Date_fin)
	if err != nil {
		return Defi{
			Num:        -1,
			Date_debut: "",
			Date_fin:   "",
		}
	}

	return defi
}

func AddDefi(num int, dateD string, dateF string) {
	stmt, err := db.Prepare("INSERT INTO Defis values(?,?,?)")
	_, err = stmt.Exec(num, dateD, dateF)
	if err != nil {
		fmt.Println("erreur add défi " + err.Error())
	}
	stmt.Close()
}
func ModifyDefi(num int, dateD string, dateF string) {
	stmt, err := db.Prepare("UPDATE Defis SET date_debut = ?, date_fin = ? where numero = ?")
	if err != nil {
		fmt.Println(err)
	}
	if _, err := stmt.Exec(dateD, dateF, num); err != nil {
		fmt.Println("erreur modify defi " + err.Error())
	}
	stmt.Close()
}
func GetResultat(login string) []ResBDD {
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
