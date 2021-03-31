package config

import (
	"strings"
	"time"
)

var (
	PathRoot       = "./"
	PathDefis      = "./ressource/defis/"
	PathScripts    = "./ressource/script_etudiants/"
	PathDirTest    = "./dir_test/"
	PathJeuDeTests = "./ressource/jeu_de_test/"
	PathLog        = "./logs/"
)

/**
 * Listes des structures utilisés
 */

// Structures a réutiliser un peu partout
type Etudiant struct {
	Login         string
	Password      string
	Prenom        string
	Nom           string
	Correcteur    bool
	ResDefiActuel []Resultat
}

func (e Etudiant) Mail() string {
	return e.Login + "@etu.univ-nantes.fr"
}

type Admin struct {
	Login    string
	Password string
}

type EtudiantMail struct {
	Login  string
	Prenom string
	Nom    string
	Defis  []ResBDD
}

func (e EtudiantMail) Mail() string {
	return e.Login + "@etu.univ-nantes.fr"
}

type ResBDD struct {
	Login     string
	Defi      int
	Etat      int
	Tentative int
}
type ParticipantDefi struct {
	Etudiant Etudiant
	Resultat ResBDD
}

type Defi struct {
	Num        int
	DateDebut  time.Time
	DateFin    time.Time
	JeuDeTest  bool
	Correcteur string
}

func (d Defi) DateDebutString() string {
	return strings.Split(d.DateDebut.String(), " ")[0]
}
func (d Defi) DateFinString() string {
	return strings.Split(d.DateFin.String(), " ")[0]
}
func (d Defi) TimeDebutString() string {
	e := strings.Split(strings.Split(d.DateDebut.String(), " ")[1], ":")
	return strings.Join([]string{e[0], e[1]}, ":")
}
func (d Defi) TimeFinString() string {
	e := strings.Split(strings.Split(d.DateDebut.String(), " ")[1], ":")
	return strings.Join([]string{e[0], e[1]}, ":")
}

// structure
type Resultat struct {
	Etat           int
	CasTest        CasTest
	Res_etu        []Retour
	Res_correction []Retour
	Error_message  string
}
type Retour struct { // changer le Nom --> dossier/fichier
	Nom     string
	Contenu string
}

type CasTest struct {
	Nom       string
	Arguments []Retour
}
