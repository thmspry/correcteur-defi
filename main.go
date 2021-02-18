package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
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
		Login:    "test",
		Password: "test",
		Prenom:   "testPrenom",
		Nom:      "testNom",
		Mail:     "testMail",
	}
	BDD.Register(etu)
}
