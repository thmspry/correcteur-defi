package testeur

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
)

var (
	path_defis      = "./defis/"
	path_script_etu = "./script_etudiants"
	path_dir_test   = "./dir_test/"
)

func Test(etudiant string) string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */

	//Sauvegarder l'arborescence du projet
	cmd := exec.Command("find", "*")
	out, err := cmd.CombinedOutput()
	if err := ioutil.WriteFile("./arboSave.txt", out, 0644); err != nil {
		fmt.Println("err d'écriture d'arboSave.txt\n", err, "\n")
	}

	script_etu := "script_" + etudiant + ".sh"
	defi := "defi_X.sh"

	if !MakeFileExecutable("./script_etudiants/" + script_etu) {
		return "chmod failed"
	}

	deplacer(path_defis+defi, path_dir_test)
	deplacer(path_script_etu+script_etu, path_dir_test)

	cmd = exec.Command("/bin/sh", script_etu)
	cmd.Dir = path_dir_test
	cmd.Env = append(os.Environ(), "USER=testeur") //ligne qui devrait permettre de faire le test en tant que UID "testeur"

	stdout_etu, err := cmd.CombinedOutput()
	if err != nil {
		return "execution de " + script_etu + " failed\n" + err.Error()
	}
	cmd = exec.Command("/bin/sh", defi)
	cmd.Dir = path_dir_test
	cmd.Env = append(os.Environ(), "USER=testeur")
	stdout_defi, err := cmd.CombinedOutput()
	if err != nil {
		return "execution de " + defi + " failed\n" + err.Error()
	}

	deplacer(path_dir_test+defi, path_defis)
	deplacer(path_dir_test+script_etu, path_script_etu)

	fmt.Printf(string(stdout_etu) + string(stdout_defi))
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
	fmt.Println(user.Current())
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
