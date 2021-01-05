package logs

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func GetDate() string {
	annee, mois, jour := time.Now().Date()
	date := strconv.Itoa(jour) + "-" + strconv.Itoa(int(mois)) + "-" + strconv.Itoa(annee)
	return date
}

func GetHoraire() string {
	minute, heure, seconde := time.Now().Clock()
	horaire := strconv.Itoa(minute) + ":" + strconv.Itoa(heure) + ":" + strconv.Itoa(seconde)
	return horaire
}

func WriteLog(login string, msg string) {
	if !testeur.Contains("./logs/", GetDate()) {
		ioutil.WriteFile("./logs/"+GetDate(), nil, 0755)
	}

	f, err := os.OpenFile("./logs/"+GetDate(), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf(err.Error())
	}
	f.WriteString(GetHoraire() + ", " + login + " : " + msg + "\n")

	f.Close()
}
