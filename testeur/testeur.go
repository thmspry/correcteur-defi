package testeur

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	path_defis        = "./ressource/defis/"
	path_script_etu   = "./ressource/script_etudiants/"
	path_dir_test     = "./dir_test/"
	path_jeu_de_tests = ""
	passTout          bool
)

func Test(etudiant string) string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */

	//Création du user
	if err := exec.Command("sudo", "useradd", etudiant).Run(); err != nil {
		fmt.Println("error create user : ", err)
	}
	if err := exec.Command("sudo", "mkhomedir_helper", etudiant).Run(); err != nil {
		fmt.Println("error create user : ", err)
	}
	path_dir_test = "/home/" + etudiant + "/"

	//Récupérer le défi actuel
	num, defi := Defi_actuel()
	script_etu := "script_" + etudiant + "_" + num + ".sh"
	path_jeu_de_tests = path_jeu_de_tests + "test_defi_" + num + "/"

	i, _ := strconv.Atoi(num)
	var resTest = make([]int, i+1) // 1 : réussi, 0 : échoué, -1 : error

	if !makeFileExecutable(path_script_etu + script_etu) {
		return "chmod failed pour" + script_etu
	}
	if !makeFileExecutable(path_defis + defi) {
		return "chmod failed pour" + defi
	}

	// Début du test_2
	tests, _ := exec.Command("find", path_jeu_de_tests, "-type", "f").CombinedOutput()
	nbJeuDeTest := len(strings.Split(string(tests), "\n")) - 1

	for i := 0; i < nbJeuDeTest; i++ {

		deplacer(path_defis+defi, path_dir_test)
		deplacer(path_script_etu+script_etu, path_dir_test)

		test := "test_" + strconv.Itoa(i)
		deplacer(path_jeu_de_tests+test, path_dir_test)

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

	if err := exec.Command("sudo", "userdel", etudiant).Run(); err != nil {
		fmt.Println("error sudo userdel : ", err)
	}
	if err := exec.Command("sudo", "rm", "-rf", "/home/"+etudiant).Run(); err != nil {
		fmt.Println("error sudo rm -rf /home/EXXX : ", err)
	}

	if passTout {
		//mettre dans la table "defis" de la BDD num etu, num defis et "réussi"
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

func testeurUnique(defi string, script_user string, etudiant string) int {

	var stderr bytes.Buffer
	var out bytes.Buffer

	arboAvant := getFiles(path_dir_test)
	makeFileExecutable(path_dir_test + defi)
	defi = "'/bin/sh " + path_dir_test + "defi'"
	cmd := exec.Command("su", etudiant, "-c", defi)
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("erreur execution script défis : ", err)
		fmt.Println(stderr.String())
		return -1
	}
	stdout_defi := out.String()
	arboApres := getFiles(path_dir_test)
	if len(arboAvant) != len(arboApres) {
		//modif dans un new fichier
		//trouver le fichier / nom du fichier modifié
		diff := difference(arboAvant, arboApres)
		mapDefi := make(map[string]string)
		mapEtu := make(map[string]string)
		for _, name := range diff {
			f, err := exec.Command("cat", path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", path_dir_test+name, "\n", err)
			}
			mapDefi[name] = string(f)
		}

		//execution script étudiant
		command := "'/bin/sh " + path_dir_test + "defi'"
		cmd := exec.Command("su", etudiant, "-c", command)
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

			if mapEtu[name] != mapDefi[name] {
				return 0
			}
		}
		return 1
	} else {
		cmd = exec.Command("/bin/sh", script_user)
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

	exec.Command("groupadd", "grpTest")
	// ajouter l'user au groupe
	exec.Command("usermod", "-a", "-G", "grpTest", "testeur")
	//empeche la modification de fichier à partir de la racine au groupe mais garder la navigabilité
	os.Chmod("./", 755)
	// empecher de créer/supprimer des fichiers du répertoire de base :
	exec.Command("chmod", "-R", "700", "./*")
	//donne le droit de modif à un dossier spécifique sur le serveur au groupe
	os.Chmod("./dir_test", 777)
	// ou chmod 770 ./testeur/dir_test
}

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
