package modele

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

/**
@Etudiant structure représentant un étudiant de la base de donnée
*/
type Etudiant struct {
	Login         string
	Password      string
	Prenom        string
	Nom           string
	Correcteur    bool
	ResDefiActuel []ResultatTest
}

/**
@Mail génère l'adresse mail de l'étudiant
*/
func (e Etudiant) Mail() string {
	return e.Login + "@etu.univ-nantes.fr"
}

/**
@Admin structure de connexion d'un compte administrateur
*/
type Admin struct {
	Login    string
	Password string
}

/**
@EtudiantMail renvoie un étudiant (login, prénom, nom) avec la liste des défis auxquelles il a répondu
*/
type EtudiantMail struct {
	Login  string
	Prenom string
	Nom    string
	Defis  []Resultat
}

func (e EtudiantMail) Mail() string {
	return e.Login + "@etu.univ-nantes.fr"
}

/**
@Resultat structure qui correspond à la table Resultat de la BDD
*/
type Resultat struct {
	Login      string
	Defi       int
	Etat       int
	Tentative  int
	Classement int
}

/**
@Defi structure qui représente un défi de la table Defi de la BDD
*/
type Defi struct {
	Num        int
	DateDebut  time.Time
	DateFin    time.Time
	JeuDeTest  bool
	Correcteur string
}

/**
@DateDebutString retourne la date de début sous le format "YYY-MM-DD"
*/
func (d Defi) DateDebutString() string {
	return strings.Split(d.DateDebut.String(), " ")[0]
}

/**
@DateFinString retourne la date de fin sous le format "YYY-MM-DD"
*/
func (d Defi) DateFinString() string {
	return strings.Split(d.DateFin.String(), " ")[0]
}

/**
@TimeDebutString retourne l'heure de début au format "HH-MM"
*/
func (d Defi) TimeDebutString() string {
	e := strings.Split(strings.Split(d.DateDebut.String(), " ")[1], ":")
	return strings.Join([]string{e[0], e[1]}, ":")
}

/**
@TimeFinString retourne l'heure de fin au format "HH-MM"
*/
func (d Defi) TimeFinString() string {
	e := strings.Split(strings.Split(d.DateFin.String(), " ")[1], ":")
	return strings.Join([]string{e[0], e[1]}, ":")
}

/**
@ParticipantDefi
*/
type ParticipantDefi struct {
	Etudiant Etudiant
	Resultat Resultat
}

// structure
type ResultatTest struct {
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

type StatsDefi struct {
	Num               int
	ParticipantsDefi  int
	Reussite          int
	MoyenneTentatives int
}

type StatsDefis struct {
	NbEtudiants  int
	Participants []StatsDefi
}
