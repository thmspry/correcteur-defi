package testeur

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

func Test(etudiant string) string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */
	script_etu := "script_" + etudiant + ".sh"
	path_dir_test := "./testeur/dir_test/"
	defi := "defi_X.sh"

	if !MakeFileExecutable("./script_etudiants/" + script_etu) {
		return "chmod failed"
	}

	deplacer("./defis/"+defi, path_dir_test)
	deplacer("./script_etudiants/"+script_etu, path_dir_test)

	cmd := exec.Command("/bin/sh", script_etu)
	//cmd.SysProcAttr = &syscall.
	//cmd.Dir = path_dir_test
	stdout_etu, err := cmd.CombinedOutput()
	if err != nil {
		return script_etu + err.Error()
	}
	cmd = exec.Command("/bin/sh", defi)
	cmd.Dir = path_dir_test
	stdout_defi, err := cmd.CombinedOutput()
	if err != nil {
		return defi + err.Error()
	}

	deplacer(path_dir_test+defi, "./defis/")
	deplacer(path_dir_test+script_etu, "./script_etudiants/")

	fmt.Printf(string(stdout_etu) + string(stdout_defi))
	if string(stdout_defi) == string(stdout_etu) {
		return "On obtient la même chose"
	} else {
		return "On obtient pas la même chose"
	}

}

func deplacer(path_in string, path_out string) bool {
	_, err := exec.Command("mv", path_in, path_out).CombinedOutput()
	if err != nil {
		fmt.Print(path_in + " non trouvé\n" + err.Error())
		return false
	}
	return true
}
func MakeFileExecutable(script string) bool {
	err := os.Chmod(script, 0755)
	if err != nil {
		fmt.Print(err.Error())
		return false
	}
	return true
}

func TestUser() {
	fmt.Println(user.Current())
}

// Non testé
func InitUser() {
	//crée l'user
	exec.Command("useradd", "testeur")
	//crée le groupe
	exec.Command("groupadd", "grpTest")
	// ajouter l'user au groupe
	exec.Command("usermod", "-a", "-G", "grpTest", "testeur")
	//empeche la modification de fichier à partir de la racine au groupe
	os.Chmod("./", 750)
	//donne le droit de modif à un dossier spécifique sur le serveur au groupe
	// ou chmod 770 ./testeur/dir_test
	exec.Command("chown", "-R", "testeur:grpTest", "testeur/dir_test")
}
