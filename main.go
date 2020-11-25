package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
)

func main() {
	//fmt.Println("testeur retourne : ", testeur.Test("EXXX"))
	web.InitWeb()
	//BDD.InitBDD()
	/*etu := BDD.Etudiant{
		Login:      "E1045",
		Password:   "3n6Z",
		Prenom:     "Paul",
		Nom:        "Vernin",
		Mail:       "paul.vernin@gmail.com",
		DefiSucess: 0,
	}*/

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
