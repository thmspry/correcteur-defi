package testeur

import (
	"fmt"
	"os"
	"os/exec"
)

func Test(etudiant string) string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */
	script_etu := "script_" + etudiant + ".sh"
	path_dir_test := "./main/testeur/dir_test/"
	defi := "defi_X.sh"

	if !MakeFileExecutable("./main/script_etudiants/" + script_etu) {
		return "chmod failed"
	}

	deplacer("./main/defis/"+defi, path_dir_test)
	deplacer("./main/script_etudiants/"+script_etu, path_dir_test)

	cmd := exec.Command("/bin/sh", script_etu)
	cmd.Dir = path_dir_test
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

	deplacer(path_dir_test+defi, "./main/defis/")
	deplacer(path_dir_test+script_etu, "./main/script_etudiants/")

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
