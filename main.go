package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
)

func main() {
	//testeur.TestUser()
	//fmt.Println(testeur.Test("EXXX"))
	//fmt.Printf(testeur.Defi_actuel())
	//fmt.Println(testeur.Test("EXXX"))
	//testeur.TesteurUnique("","")
	/*	BDD.InitBDD()

		etu := BDD.Etudiant{
			Login:      "test",
			Password:   "test",
			Prenom:     "testPrenom",
			Nom:        "testNom",
			Mail:       "testMail",
			DefiSucess: 0,
		}
		BDD.Register(etu)
	*/
	web.InitWeb()

	//testeur.Test("EXXX")

}
