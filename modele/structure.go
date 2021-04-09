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
@ParticipantDefi concatene toutes les informations d'un participants à un défi (prénom, nom, état, tentative, etc..)
*/
type ParticipantDefi struct {
	Etudiant Etudiant
	Resultat Resultat
}

type ResultatTest struct {
	Etat           int      // 1 : réussi, 0 : raté, -1 : erreur
	CasTest        CasTest  // cas de test du jeu de test
	Res_etu        []Retour // résultat obtenu par l'étudiant
	Res_correction []Retour // résultat obtenu par la correction
	Error_message  string   // "" si pas d'erreur, sinon message d'erreur obtenu lors du test
}

/**
@Retour structure qui représenter à la fois les fichiers et les dossiers
*/
type Retour struct {
	Nom     string
	Contenu string
}

type CasTest struct {
	Nom       string
	Arguments []Retour
}

/* --- Statistiques ---*/
type StatsDefi struct {
	Num               int
	ParticipantsDefi  int
	Reussite          int
	MoyenneTentatives int
}

type StatsDefis struct {
	NbEtudiants  int
	NbDefiActuel int
	Participants []StatsDefi
}
