package testeur

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var (
	Path_defis        = "./ressource/defis/"
	Path_script_etu   = "./ressource/script_etudiants/"
	Path_dir_test     = "./dir_test/"
	Path_jeu_de_tests = "./ressource/jeu_de_test/"
	passTout          bool
)

type Resultat struct {
	Etat           int
	Test           string
	Res_etu        []Retour
	Res_correction []Retour
	Error_message  string
}
type Retour struct {
	Nom     string
	Contenu string
}

/*
Fonction a appelé pour tester
*/
func Test(etudiant string) (string, []Resultat) {

	os := runtime.GOOS
	if os == "windows" || os == "darwin" {
		return "Le testeur ne peut être lancé que sur linux", nil
	}

	//Création du user
	if err := exec.Command("useradd", etudiant).Run(); err != nil {
		fmt.Println("error create user : ", err)
	}
	//Associe le dossier à l'user
	if err := exec.Command("mkhomedir_helper", etudiant).Run(); err != nil {
		fmt.Println("error create dir : ", err)
	}
	//Associe le chemin de Test au dossier créé
	Path_dir_test = "/home/" + etudiant + "/"
	if err := exec.Command("chmod", "770", Path_dir_test).Run(); err != nil {
		fmt.Println("error chmod "+Path_dir_test, err)
	}
	clear(Path_dir_test)

	//Récupérer le défi actuel
	num_defi, defi := Defi_actuel()
	script_etu := "script_" + etudiant + "_" + strconv.Itoa(num_defi) + ".sh"
	Path_jeu_de_tests = Path_jeu_de_tests + "test_defi_" + strconv.Itoa(num_defi) + "/"

	var resTest = make([]Resultat, 0)

	// Début du testUnique

	// Donne les droits d'accès et de modifications à l'étudiant
	exec.Command("sudo", "chown", etudiant, Path_script_etu+script_etu).Run()
	exec.Command("sudo", "chown", etudiant, Path_dir_test).Run()

	//On récupère le nombre de Test pour faire la boucle
	tests, _ := exec.Command("find", Path_jeu_de_tests, "-type", "f").CombinedOutput()
	nbJeuDeTest := len(strings.Split(string(tests), "\n")) - 1

	for i := 0; i < nbJeuDeTest; i++ {
		deplacer(Path_defis+defi, Path_dir_test)
		deplacer(Path_script_etu+script_etu, Path_dir_test)
		test := "test_" + strconv.Itoa(i)
		deplacer(Path_jeu_de_tests+test, Path_dir_test)
		exec.Command("chmod", "444", Path_dir_test+test).Run()
		rename(Path_dir_test, test, "Test")

		res := testeurUnique(defi, script_etu, etudiant)
		f, _ := ioutil.ReadFile(Path_dir_test + test)
		res.Test = string(f)
		resTest = append(resTest, res)

		rename(Path_dir_test, "Test", test)
		deplacer(Path_dir_test+test, Path_jeu_de_tests)
		deplacer(Path_dir_test+defi, Path_defis)
		deplacer(Path_dir_test+script_etu, Path_script_etu)
		clear(Path_dir_test)

		if res.Etat == 0 {
			BDD.SaveDefi(etudiant, num_defi, 0, false)
			return "Le Test n°" + strconv.Itoa(i) + " n'est pas passé", resTest
		}
		if res.Etat == -1 {
			BDD.SaveDefi(etudiant, num_defi, 0, false)
			return "Il y a eu un erreur lors du Test n°" + strconv.Itoa(i), resTest
		}
	}

	clear(Path_dir_test)

	//supprime l'user et son dossier
	if err := exec.Command("sudo", "userdel", etudiant).Run(); err != nil {
		fmt.Println("error sudo userdel : ", err)
	}
	if err := exec.Command("sudo", "rm", "-rf", "/home/"+etudiant).Run(); err != nil {
		fmt.Println("error sudo rm -rf /home/EXXX : ", err)
	}

	BDD.SaveDefi(etudiant, num_defi, 1, false)

	return "Vous avez passé tous les tests avec succès", resTest
}

/*
Fonctionnement du testeur unique :
- execute le script du defis
- stock le résultat de celui-ci
- regarde si il y a eu des nouveaux fichiers qui ont été crée ou non
Si oui :
	- stock le contenu des nouveaux fichiers et leurs nom dans une map
	- lance le script etu
	- stock le contenu des fichiers et leurs noms dans une map étudiant
	- compare le contenu des fichiers
	- return 1 si c'est pareil, 0 sinon
Si non :
	- lance le script étu
	- stock son retour imédiat (stdout)
	- compare stdout du défis et de l'étudiant
	- retunr 1 si c'est pareil, 0 sinon
*/
func testeurUnique(defi string, script_user string, etudiant string) Resultat {

	res := Resultat{
		Etat:           0,
		Res_etu:        make([]Retour, 0),
		Res_correction: make([]Retour, 0),
		Error_message:  "",
	}
	retour := Retour{
		Nom:     "",
		Contenu: "",
	}
	arboAvant := getFiles(Path_dir_test)

	cmd := exec.Command(Path_dir_test + defi)
	cmd.Dir = Path_dir_test
	stdout_defi, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("erreur execution script défis : ", err)
		res.Error_message = "erreur execution du script de correction"
		res.Etat = -1
		return res
	}
	arboApres := getFiles(Path_dir_test)
	if len(arboAvant) != len(arboApres) {
		//modif dans un new fichier
		//trouver le fichier / nom du fichier modifié
		diff := difference(arboAvant, arboApres)
		mapDefi := make(map[string]string)
		mapEtu := make(map[string]string)
		for _, name := range diff {
			//On donne seulement le droit de lecture sur les jeux de Test
			exec.Command("chmod", "444", Path_dir_test+name).Run()
			f, err := exec.Command("cat", Path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", Path_dir_test+name, "\n", err)
				res.Error_message = "erreur lecture du fichier " + Path_dir_test + name
				res.Etat = -1
				return res
			}
			mapDefi[name] = string(f)
			retour.Nom = "fichier " + name
			retour.Contenu = string(f)

			res.Res_correction = append(res.Res_correction, retour)
		}
		//execution script étudiant

		command := "'" + Path_dir_test + script_user + "'"
		cmd := exec.Command("sudo", "-H", "-u", etudiant, "bash", "-c", command)
		cmd.Dir = Path_dir_test
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("erreur execution script étudiant : ", err)
			res.Error_message = "erreur execution du script de l'étudiant"
			res.Etat = -1
			return res
		}
		//Récup les fichiers
		for _, name := range diff {
			f, err := exec.Command("cat", Path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", Path_dir_test+name, "\n", err)
				res.Error_message = "erreur lecture du fichier " + Path_dir_test + name
				res.Etat = -1
				return res
			}
			mapEtu[name] = string(f)
			retour.Nom = "fichier " + name
			retour.Contenu = string(f)
			res.Res_etu = append(res.Res_etu, retour)
			if mapEtu[name] != mapDefi[name] {
				res.Etat = 0
				return res
			}
		}
		res.Etat = 1
		return res
	} else {
		command := "'" + Path_dir_test + script_user + "'"
		cmd := exec.Command("sudo", "-H", "-u", etudiant, "bash", "-c", command)
		cmd.Dir = Path_dir_test
		stdout_etu, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err, stdout_etu)
			res.Error_message = "erreur execution du script étudiant"
			res.Etat = -1
			return res
		}

		retour.Nom = "sortie standart"
		retour.Contenu = string(stdout_etu)
		res.Res_etu = append(res.Res_etu, retour)
		retour.Contenu = string(stdout_defi)
		res.Res_correction = append(res.Res_correction, retour)
		if res.Res_etu[0] == res.Res_correction[0] {
			res.Etat = 1
			return res
		} else {
			res.Etat = 0
			return res
		}
	}
}

// testé à la main mais pas avec go sur le serveur
func InitUser() {

}

/*
Fonction qui renvoie un tableau contenant la différence entre les deux tableaux entrés en paramètre
*/
func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return diff
}
