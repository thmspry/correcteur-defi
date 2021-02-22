package logs

import (
	"fmt"
	date "github.com/aodin/date"
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

func WriteLog(titre string, msg string) {
	d := date.Today()
	/*if !manipStockage.Contains("./logs/", d.String()) {
		ioutil.WriteFile("./logs/"+d.String(), nil, 0755)
	}*/
	_, err := os.Stat("./logs/" + d.String())
	if os.IsNotExist(err) {
		ioutil.WriteFile("./logs/"+d.String(), nil, 0755)
	}

	f, err := os.OpenFile("./logs/"+d.String(), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf(err.Error())
	}
	f.WriteString(GetHoraire() + ", " + titre + " : " + msg + "\n")

	f.Close()
}
