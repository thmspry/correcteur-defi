package BDD

import (
	"database/sql"
	"fmt"
	"time"
)

// Structure a réutiliser un peu partout
type Etudiant struct {
	Login      string
	Password   string
	Prenom     string
	Nom        string
	Mail       string
	DefiSucess int
}

var db, _ = sql.Open("sqlite3", "./BDD/projS3.db")

func InitBDD() {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS Etudiant (" +
		"login TEXT PRIMARY KEY, " +
		"password TEXT NOT NULL, " +
		"prenom TEXT NOT NULL," +
		"nom TEXT NOT NULL," +
		"mail TEXT NOT NULL," +
		"defiSucess INTEGER NOT NULL" +
		");")
	if err != nil {
		fmt.Println("prblm table Etudiant" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Defis (" +
		"login TEXT NOT NULL," +
		"defi INTEGER NOT NULL," +
		"etat TEXT NOT NULL," + // 3 états : R (réussi), T (tenté), N (non tenté):om
		"FOREIGN KEY (login) REFERENCES Etudiant(login)" +
		")")
	if err != nil {
		fmt.Println("prblm table Defis" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Token (" +
		"login TEXT NOT NULL PRIMARY KEY," +
		"token TEXT NOT NULL," +
		"FOREIGN KEY(login) REFERENCES Etudiant(login)" +
		")")
	if err != nil {
		fmt.Println("Erreur dans la table Token" + err.Error())
	}
	stmt.Exec()

}

//testé
func Register(etu Etudiant) bool {
	stmt, err := db.Prepare("INSERT INTO Etudiant values(?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
	}

	_, err = stmt.Exec(etu.Login, etu.Password, etu.Prenom, etu.Nom, etu.Mail, etu.DefiSucess)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("l'étudiant de login : " + string(etu.Login) + " a été enregistré dans la bdd\n")
	stmt.Close()
	return true
}

//testé
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

//testé
func GetInfo(id string) Etudiant {
	fmt.Println("fonc GetInfo : ")
	var (
		login      string
		password   string
		prenom     string
		nom        string
		mail       string
		defiSucess int
	)

	row := db.QueryRow("SELECT * FROM Etudiant WHERE login = $1", id)
	err := row.Scan(&login, &password, &prenom, &nom, &mail, &defiSucess)

	if err != nil {
		fmt.Printf("problme row scan \n", err)
	} else {
		fmt.Println("etu : ", login, password, prenom, nom, mail, defiSucess)
	}
	etu := Etudiant{
		Login:      login,
		Password:   password,
		Prenom:     prenom,
		Nom:        nom,
		Mail:       mail,
		DefiSucess: defiSucess,
	}

	fmt.Println("/ fonc GetInfo")
	return etu
}

func GetNameByToken(token string) string {
	var login string
	row := db.QueryRow("SELECT * FROM token WHERE token = $1", token)
	err := row.Scan(&login, &token)
	if err != nil {
		fmt.Printf("problme row scan \n", err)
	}
	return login
}

func InsertToken(login string, token string) {
	stmt, err := db.Prepare("INSERT INTO Token values(?,?)")
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmt.Exec(login, token)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Ajout token : ", token, " pour ", login, " à ", time.Now(), "\n")
	}
	stmt.Close()
}

func DeleteToken(token string) {
	stmt, _ := db.Prepare("DELETE FROM token WHERE token = ?")
	_, err := stmt.Exec(token)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("delete token : ", token, " à ", time.Now(), "\n")
	}
	stmt.Close()
}
