package logs

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func GetHoraire() string {
	minute, heure, seconde := time.Now().Clock()
	horaire := strconv.Itoa(minute) + ":" + strconv.Itoa(heure) + ":" + strconv.Itoa(seconde)
	return horaire
}

func WriteLog(login string, msg string) {
	date := testeur.GetDate()
	if !testeur.Contains("./logs/", date.String) {
		ioutil.WriteFile("./logs/"+date.String, nil, 0755)
	}

	f, err := os.OpenFile("./logs/"+date.String, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf(err.Error())
	}
	f.WriteString(GetHoraire() + ", " + login + " : " + msg + "\n")

	f.Close()
}
