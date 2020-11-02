package testeur

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Test() string {

	if !MakeFileExecutable() {
		return "Fichier non trouvé"
	}

	cmd := exec.Command("/bin/sh", "script.sh")
	cmd.Dir = "./main/testeur/"
	stdout, err := cmd.CombinedOutput()

	fmt.Print(string(stdout))
	if err != nil {
		fmt.Print("cmd.Run() de Test() failed with \n", err, "\n")
		return err.Error()
	}

	return string(stdout)
}

func MakeFileExecutable() bool {
	err := os.Chmod("main/testeur/script.sh", 0755)
	if err != nil {
		log.Fatalf("chmod failed with \n", err)
		return false
	}
	return true
}
