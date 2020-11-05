package testeur

import (
	"fmt"
	"os"
	"os/exec"
)

func Test() string {
	/*
	* Sauvegarder le layout avant execution du script
	* executer le script dans un dossier sans pouvoir revenir plus haut et faire cd fait venir dans ce dossier
	 */
	if !MakeFileExecutable("script_EXXX.sh") {
		return "chmod failed"
	}

	cmd := exec.Command("/bin/sh", "script_EXXX.sh")
	cmd.Dir = "./main/testeur/"
	stdout, err := cmd.CombinedOutput()

	fmt.Print(string(stdout))
	if err != nil {
		fmt.Print("cmd.Run() de Test() failed with \n", err, "\n")
		return err.Error()
	}

	return string(stdout)
}

func MakeFileExecutable(script string) bool {
	err := os.Chmod("main/testeur/"+script, 0755)
	if err != nil {
		return false
	}
	return true
}
