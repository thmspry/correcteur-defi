package logs

import (
	"fmt"
	date "github.com/aodin/date"
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
	d := date.Today()
	if !testeur.Contains("./logs/", d.String()) {
		ioutil.WriteFile("./logs/"+d.String(), nil, 0755)
	}

	f, err := os.OpenFile("./logs/"+d.String(), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf(err.Error())
	}
	f.WriteString(GetHoraire() + ", " + login + " : " + msg + "\n")

	f.Close()
}
