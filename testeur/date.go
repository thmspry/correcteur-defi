package testeur

import (
	"strconv"
	"strings"
	"time"
)

type Date struct {
	jour   int
	mois   int
	annee  int
	String string
}

func GetDate() Date {
	annee, mois, jour := time.Now().Date()
	date := Date{
		jour:   jour,
		mois:   int(mois),
		annee:  annee,
		String: strconv.Itoa(jour) + "-" + strconv.Itoa(int(mois)) + "-" + strconv.Itoa(annee),
	}
	return date
}
func GetDateFromString(date string) Date {
	d := strings.Split(date, "-")
	j, _ := strconv.Atoi(d[0])
	m, _ := strconv.Atoi(d[1])
	a, _ := strconv.Atoi(d[2])
	res := Date{
		jour:   j,
		mois:   m,
		annee:  a,
		String: date,
	}
	return res
}
func DatePassed(date1 Date) bool {
	date2 := GetDate()
	if date1.annee > date2.annee {
		return false
	} else if date1.annee < date2.annee {
		return true
	}
	if date1.mois > date2.mois {
		return false
	} else if date1.mois < date2.mois {
		return true
	}
	if date1.jour > date2.jour {
		return false
	} else if date1.jour < date2.jour {
		return true
	}
	return false
}
