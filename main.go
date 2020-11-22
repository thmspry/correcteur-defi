package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
)

type Etudiant struct {
	Login    string
	Password string
	Prenom   string
	Nom      string
}

func main() {
	//fmt.Println("testeur retourne : ", testeur.Test("EXXX"))
	//web.InitWeb()
	BDD.InitBDD()

	/*
		rows, _ := db.Query("SELECT id, firstname FROM etudiant")
		stmt.Exec("ouisqd")
		var id int
		var firstname string
		for rows.Next() {
			rows.Scan(&id, &firstname)
			fmt.Printf(strconv.Itoa(id) + ": "+ firstname + "\n")
		}*/
}
