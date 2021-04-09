package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	web "gitlab.univ-nantes.fr/E192543L/projet-s3/modele/web"
	"os"
)

//Main fonction du programme
func main() {
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
		//permet d'exécuter des tests unitaires
	} else {
		web.InitWeb()
	}
}

/**
@reset fonction pour reset le contenu des dossiers et de la base de donnée
*/
func reset() {
	if manipStockage.Contains("./DAO/", "database.db") {
		os.Remove("./DAO/database.db")
	}
	DAO.InitDAO()
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

/**
@Init fonction pour initialiser le serveur et les différents répertoires
*/
func Init() {
	os.Mkdir("./logs", 0755)
	os.Mkdir("./DAO", 0755)

	path := modele.PathRoot + "/ressource"
	os.Mkdir(path, 0755)
	os.Mkdir(path+"/defis", 0755)
	os.Mkdir(path+"/script_etudiants", 0755)
	os.Mkdir(path+"/jeu_de_test", 0755)
	DAO.InitDAO()
}
