package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

/**
@GetHoraire retourne l'horaire sous le format HH:MM:SS
*/
func GetHoraire() string {
	minute, heure, seconde := time.Now().Clock()
	horaire := strconv.Itoa(minute) + ":" + strconv.Itoa(heure) + ":" + strconv.Itoa(seconde)
	return horaire
}

/**
@WriteLog permet d'écrire des logs de ce qu'il s'est passé sur le serveur
*/
func WriteLog(titre string, msg string) {
	d := strings.Split(time.Now().String(), " ")[0]
	_, err := os.Stat("./logs/" + d)
	if os.IsNotExist(err) {
		ioutil.WriteFile("./logs/"+d, nil, 0755)
	}

	f, err := os.OpenFile("./logs/"+d, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf(err.Error())
	}
	f.WriteString(GetHoraire() + ", " + titre + " : " + msg + "\n")

	f.Close()
}
