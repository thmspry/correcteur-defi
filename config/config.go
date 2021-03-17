package config

import (
	"github.com/aodin/date"
)

var (
	Path_root         = "./"
	Path_defis        = "./ressource/defis/"
	Path_scripts      = "./ressource/script_etudiants/"
	Path_dir_test     = "./dir_test/"
	Path_jeu_de_tests = "./ressource/jeu_de_test/"
	Path_log          = "./logs/"
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
	DateDebut  date.Date
	DateFin    date.Date
	JeuDeTest  bool
	Correcteur string
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
