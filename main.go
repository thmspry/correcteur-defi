package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
	"os"
)

//Main fonction du programme
func main() {

	// initialalise le routeur

	/*
		set GOOS=linux
		set GOARCH=amd64
	*/
	mode := os.Args[1]
	if mode == "init" {
		fmt.Println("init")
		Init()
	} else if mode == "reset" {
		fmt.Println("reset")
		reset()
	} else if mode == "start" {
		fmt.Println("start")
		web.InitWeb()
	} else if mode == "test" {
		fmt.Println(BDD.GetResultatActuel("E197051L"))
	} else {
		web.InitWeb()
	}

}

//fonction pour reset les dossiers et la bdd
func reset() {
	if manipStockage.Contains("./BDD/", "database.db") {
		os.Remove("./BDD/database.db")
	}
	BDD.InitBDD()
	if len(manipStockage.GetFiles("./logs")) > 0 {
		os.RemoveAll("./logs")
		os.Mkdir("./logs", 0755)
	}
	if len(manipStockage.GetFiles("./ressource/defis")) > 0 {
		os.RemoveAll("./ressource/defis")
		os.Mkdir("./ressource/defis", 0755)

	}
	if len(manipStockage.GetFiles("./ressource/jeu_de_test")) > 0 {
		os.RemoveAll("./ressource/jeu_de_test")
		os.Mkdir("./ressource/jeu_de_test", 0755)

	}
	if len(manipStockage.GetFiles("./ressource/script_etudiants")) > 0 {
		os.RemoveAll("./ressource/script_etudiants")
		os.Mkdir("./ressource/script_etudiants", 0755)

	}
}

// fonction pour initialiser le serveur et les différents fichiers
func Init() {
	os.Mkdir("./logs", 0755)
	os.Mkdir("./BDD", 0755)

	path := config.Path_root + "/ressource"
	os.Mkdir(path, 0755)
	os.Mkdir(path+"/defis", 0755)
	os.Mkdir(path+"/script_etudiants", 0755)
	os.Mkdir(path+"/jeu_de_test", 0755)

	BDD.InitBDD()
}
