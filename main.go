package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
)

func main() {
	//fmt.Println("testeur retourne : ", testeur.Test("EXXX"))
	/*
		étudiant test_2 :
		1 : E1045, 3n6Z
		2 : E1000, E1000


	*/

	//testeur.TestUser()
	//fmt.Println(testeur.Test("EXXX"))
	//fmt.Printf(testeur.Defi_actuel())
	fmt.Println(testeur.Test("EXXX"))
	//testeur.TesteurUnique("","")
	//web.InitWeb()
	//BDD.InitBDD()

}
