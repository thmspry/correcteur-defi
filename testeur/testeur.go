package testeur

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var (
	path_defis        = "./ressource/defis/"
	path_script_etu   = "./ressource/script_etudiants/"
	path_dir_test     = "./dir_test/"
	path_jeu_de_tests = "./ressource/jeu_de_test/"
	passTout          bool
)

/*
Fonction a appelé pour tester
*/
func Test(etudiant string) string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */

	//Création du user
	if err := exec.Command("useradd", etudiant).Run(); err != nil {
		fmt.Println("error create user : ", err)
	}
	//Associe le dossier à l'user
	if err := exec.Command("mkhomedir_helper", etudiant).Run(); err != nil {
		fmt.Println("error create dir : ", err)
	}
	//Associe le chemin de test au dossier créé
	path_dir_test = "/home/" + etudiant + "/"
	if err := exec.Command("chmod", "770", path_dir_test).Run(); err != nil {
		fmt.Println("error chmod "+path_dir_test, err)
	}
	clear(path_dir_test)

	//Récupérer le défi actuel
	num, defi := Defi_actuel()
	script_etu := "script_" + etudiant + "_" + num + ".sh"
	path_jeu_de_tests = path_jeu_de_tests + "test_defi_" + num + "/"
	i, _ := strconv.Atoi(num)
	//crée un tableau de int avec 1 = réussi, 0 = échoué, -1 = erreur
	var resTest = make([]int, i+1)

	// Début du testUnique

	// Donne les droits d'accès et de modifications à l'étudiant
	exec.Command("sudo", "chown", etudiant, path_script_etu+script_etu).Run()
	exec.Command("sudo", "chown", etudiant, path_dir_test).Run()

	//On récupère le nombre de test pour faire la boucle
	tests, _ := exec.Command("find", path_jeu_de_tests, "-type", "f").CombinedOutput()
	nbJeuDeTest := len(strings.Split(string(tests), "\n")) - 1

	for i := 0; i < nbJeuDeTest; i++ {

		deplacer(path_defis+defi, path_dir_test)
		deplacer(path_script_etu+script_etu, path_dir_test)

		test := "test_" + strconv.Itoa(i)
		deplacer(path_jeu_de_tests+test, path_dir_test)
		exec.Command("chmod", "777", path_dir_test+test).Run()

		rename(path_dir_test, test, "test")

		resTest[i] = testeurUnique(defi, script_etu, etudiant)
		if resTest[i] == 0 || resTest[i] == -1 {
			passTout = false
		}

		rename(path_dir_test, "test", test)
		deplacer(path_dir_test+test, path_jeu_de_tests)
		deplacer(path_dir_test+defi, path_defis)
		deplacer(path_dir_test+script_etu, path_script_etu)
		clear(path_dir_test)
	}

	clear(path_dir_test)

	//supprime l'user et son dossier
	if err := exec.Command("sudo", "userdel", etudiant).Run(); err != nil {
		fmt.Println("error sudo userdel : ", err)
	}
	if err := exec.Command("sudo", "rm", "-rf", "/home/"+etudiant).Run(); err != nil {
		fmt.Println("error sudo rm -rf /home/EXXX : ", err)
	}

	if passTout {
		//TODO mettre dans la table "defis" de la BDD num etu, num defis et "réussi"
	}

	res := ""
	for i := 0; i < len(resTest); i++ {
		if resTest[i] == 1 {
			res = res + "Test N°" + strconv.Itoa(i) + " : réussi\n"
		} else if resTest[i] == 0 {
			res = res + "Test N°" + strconv.Itoa(i) + " : échoué\n"
		} else {
			res = res + "Test N°" + strconv.Itoa(i) + " : error\n"
		}
	}
	return res
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
func testeurUnique(defi string, script_user string, etudiant string) int {

	arboAvant := getFiles(path_dir_test)

	cmd := exec.Command(path_dir_test + defi)
	cmd.Dir = path_dir_test
	stdout_defi, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("erreur execution script défis : ", err)
		return -1
	}
	arboApres := getFiles(path_dir_test)
	if len(arboAvant) != len(arboApres) {
		//modif dans un new fichier
		//trouver le fichier / nom du fichier modifié
		diff := difference(arboAvant, arboApres)
		mapDefi := make(map[string]string)
		mapEtu := make(map[string]string)
		for _, name := range diff {
			//On donne seulement le droit de lecteur sur les jeux de test
			exec.Command("chmod", "444", path_dir_test+name).Run()
			f, err := exec.Command("cat", path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", path_dir_test+name, "\n", err)
				return -1
			}
			mapDefi[name] = string(f)
		}
		//execution script étudiant

		command := "'" + path_dir_test + script_user + "'"
		cmd := exec.Command("sudo", "-H", "-u", etudiant, "bash", "-c", command)
		cmd.Dir = path_dir_test
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("erreur execution script étudiant : ", err)
			return -1
		}
		//Récup les fichiers
		for _, name := range diff {
			f, err := exec.Command("cat", path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", path_dir_test+name, "\n", err)
				return -1
			}
			mapEtu[name] = string(f)
			fmt.Println("mapEtu[" + name + "] = " + mapEtu[name])
			fmt.Println("mapDefi[" + name + "] = " + mapDefi[name])
			if mapEtu[name] != mapDefi[name] {
				return 0
			}
		}
		return 1
	} else {
		command := "'" + path_dir_test + script_user + "'"
		cmd := exec.Command("sudo", "-H", "-u", etudiant, "bash", "-c", command)
		cmd.Dir = path_dir_test
		stdout_etu, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err, stdout_etu)
			return -1
		}

		if string(stdout_defi) == string(stdout_etu) {
			return 1
		} else {
			return 0
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
