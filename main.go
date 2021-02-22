package main

import (
	fmt "fmt"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
	"golang.org/x/crypto/bcrypt"
	"os"
)

func main() {

	web.InitWeb()
	/*
		set GOOS=linux
		set GOARCH=amd64
	*/
	/*
		mode := os.Args[1]
		if mode == "0" {
			Init()
		} else if mode == "1" {
			web.InitWeb()
		} else {
			web.InitWeb()
		}*/

}

func Init() {
	os.Mkdir("./logs", 0755)
	os.Mkdir("./BDD", 0755)

	path := config.Path_root + "/ressource"
	os.Mkdir(path, 0755)
	os.Mkdir(path+"/defis", 0755)
	os.Mkdir(path+"/script_etudiants", 0755)
	os.Mkdir(path+"/jeu_de_test", 0755)

	resetBDD()
}

func resetBDD() {
	if manipStockage.Contains("./BDD/", "database.db") {
		os.Remove("./BDD/database.db")
	}

	BDD.InitBDD()
	etu := BDD.Etudiant{
		Login:      "test",
		Password:   "test",
		Prenom:     "testPrenom",
		Nom:        "testNom",
		Mail:       "testMail",
		Correcteur: false,
	}
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(etu.Password), 14) // hashage du mot de passe
	if err == nil {
		etu.Password = string(passwordHashed)
	}
	BDD.Register(etu)

	etu2 := BDD.Etudiant{
		Login:      "test2",
		Password:   "test2",
		Prenom:     "testPrenom2",
		Nom:        "testNom2",
		Mail:       "testMail2",
		Correcteur: false,
	}
	BDD.Register(etu2)
	BDD.SaveResultat("test", 1, 1, false)
	BDD.SaveResultat("test2", 1, 1, false)
	admin := BDD.Admin{
		Login:    "admin",
		Password: "admin",
	}

	BDD.RegisterAdmin(admin)

}
