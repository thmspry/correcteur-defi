package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
)

func main() {

	BDD.InitBDD()
	etu := BDD.Etudiant{
		Login:      "test",
		Password:   "test",
		Prenom:     "testPrenom",
		Nom:        "testNom",
		Mail:       "testMail",
		DefiSucess: 0,
	}
	BDD.Register(etu)

	//web.InitWeb()
}
