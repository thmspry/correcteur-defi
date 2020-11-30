package testeur

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var (
	path_defis        = "./ressource/defis/"
	path_script_etu   = "./ressource/script_etudiants/"
	path_dir_test     = "./dir_test/"
	path_jeu_de_tests = "./ressource/jeu_de_test/"
)

type defi struct {
	nom  string
	etat int
}

func Test(etudiant string) string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */

	num, defi := Defi_actuel()
	script_etu := "script_" + etudiant + "_" + num + ".sh"
	path_jeu_de_tests = path_jeu_de_tests + "test_defi_" + num + "/"

	i, _ := strconv.Atoi(num)
	var resTest = make([]int, i) // 1 : réussi, 0 : échoué, -1 : error
	fmt.Println(resTest)

	if !MakeFileExecutable(path_script_etu + script_etu) {
		return "chmod failed"
	}

	deplacer(path_defis+defi, path_dir_test)
	deplacer(path_script_etu+script_etu, path_dir_test)

	// Début du test
	tests, _ := exec.Command("find", path_jeu_de_tests, "-type", "f").CombinedOutput()
	nbJeuDeTest := len(strings.Split(string(tests), "\n")) - 1

	for i := 0; i < nbJeuDeTest; i++ {
		test := "test_" + string(i)
		deplacer(path_jeu_de_tests+test, path_dir_test)

		//test...
		//Recup le résultat du defi du prof

		deplacer(path_dir_test+test, path_jeu_de_tests)
	}

	cmd := exec.Command("./" + script_etu)
	cmd.Dir = path_dir_test
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: 501, Gid: 20}
	//ça passe jusqu'ici
	stdout_etu, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("execution de "+script_etu+" failed\n", err)
	}
	cmd = exec.Command("/bin/sh", defi)
	cmd.Dir = path_dir_test
	//fmt.Print("cred du new cmd : ", cmd.SysProcAttr.Credential)
	stdout_defi, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("execution de "+defi+" failed\n", err)
	}

	deplacer(path_dir_test+defi, path_defis)
	deplacer(path_dir_test+script_etu, path_script_etu)

	clear(path_dir_test)

	fmt.Printf("\nstdout_etu : " + string(stdout_etu) + "stdout_defi : " + string(stdout_defi))
	if string(stdout_defi) == string(stdout_etu) {
		return "On obtient la même chose"
	} else {
		return "On obtient pas la même chose"
	}

}

// fonction qui déplace un fichier (en ayant précisé son chemin pour le trouver) dans un nouveau dossier
func deplacer(file string, path_out string) bool {
	if _, err := exec.Command("mv", file, path_out).CombinedOutput(); err != nil {
		fmt.Println(file, " not found")
		return false
	}
	return true
}

// fonction qui rend le fichier executable
func MakeFileExecutable(script string) bool {
	if err := os.Chmod(script, 0755); err != nil {
		fmt.Print("chmod on ", script, " failed")
		return false
	}
	return true
}

func TestUser() {
	/*
		usr, _ := user.Lookup("Paul")
		fmt.Print(usr.Uid)
		fmt.Println(usr.Gid)
		fmt.Println(usr)
	*/
}

// testé à la main mais pas avec go sur le serveur
func InitUser() {
	//crée l'user
	exec.Command("useradd", "testeur")
	//crée le groupe
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
