package BDD

import (
	"database/sql"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
	_ "gitlab.univ-nantes.fr/E192543L/projet-s3/web"
)

var db, _ = sql.Open("sqlite3", "./BDD/projS3.db")

func InitBDD() {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS Etudiant (" +
		"login TEXT PRIMARY KEY, " +
		"password TEXT NOT NULL, " +
		"prenom TEXT NOT NULL," +
		"nom TEXT NOT NULL," +
		"mail TEXT NOT NULL" +
		"defiSucess INTEGER" +
		"" +
		");")
	if err != nil {
		fmt.Println("prblm table Etudiant" + err.Error())
	}
	stmt.Exec()

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS Defs (" +
		"login TEXT," +
		"defi INTEGER," +
		"etat TEXT" +
		"FOREIGN KEY (login) REFERENCES Etudiant(login)" +
		")")
	if err != nil {
		fmt.Println("prblm table Defis" + err.Error())
	}
	stmt.Exec()
}

func Register() {
	stmt, err := db.Prepare(" INTO Etudiant values(?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
	}
	res, err := stmt.Exec("", "", "", "", "", 0)
	if err != nil {
		fmt.Println(err)
	}
	id, err := res.LastInsertId()
	fmt.Printf(string(id))
}

func LoginCorrect(id string, password string) web.Etudiant {
	stmt, err := db.Prepare("SELECT * FROM Etudiant WHERE login = ? AND password = ?")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(id, password)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res.RowsAffected())
	etu := web.Etudiant{id, password, "", "", "", 0}
	if res != nil {
		return etu
	}
	return etu
}
