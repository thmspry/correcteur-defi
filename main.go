package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
)

func main() {
	/*cmd := exec.Command( "./mscript.sh", "-E", "seccomp.enabled=false")
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: 501, Gid: 20}
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("erreur execution script défis : ", err)
	}*/
	//fmt.Println("testeur retourne : ", testeur.Test("EXXX"))
	/*
		étudiant test_2 :
		1 : E1045, 3n6Z
		2 : E1000, E1000
	*/

	//testeur.TestUser()
	//fmt.Println(testeur.Test("EXXX"))
	//fmt.Printf(testeur.Defi_actuel())
	//fmt.Println(testeur.Test("EXXX"))
	//testeur.TesteurUnique("","")
	web.InitWeb()
	//BDD.InitBDD()

	/* Etudiant de test :
	etu := BDD.Etudiant{
		Login:      "test",
		Password:   "test",
		Prenom:     "testPrenom",
		Nom:        "testNom",
		Mail:       "testMail",
		DefiSucess: 0,
	}
	BDD.Register(etu)*/
	//testeur.Test("EXXX")

}
