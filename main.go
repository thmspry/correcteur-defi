package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/web"
	"os"
	"strings"
	"time"
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
		t := time.Now()
		fmt.Println(strings.Split(t.String(), " ")[0])
	} else {
		web.InitWeb()
	}

}

//fonction pour reset les dossiers et la dao
func reset() {
	if manipStockage.Contains("./DAO/", "database.db") {
		os.Remove("./DAO/database.db")
	}
	DAO.InitDAO()
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
	os.Mkdir("./DAO", 0755)

	path := modele.PathRoot + "/ressource"
	os.Mkdir(path, 0755)
	os.Mkdir(path+"/defis", 0755)
	os.Mkdir(path+"/script_etudiants", 0755)
	os.Mkdir(path+"/jeu_de_test", 0755)

	DAO.InitDAO()
}
