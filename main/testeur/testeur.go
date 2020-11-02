package testeur

import (
	"log"
	"os"
	"os/exec"
	"runtime"
)

func Test() string {

	if !MakeFileExecutable() {
		return "Fichier non trouvé"
	}

	cmd := exec.Command("/bin/sh", "script.sh")
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return err.Error()
	}

	return string(stdout)
}

func MakeFileExecutable() bool {
	cmd := exec.Command("ls", "-la")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	}
	out, err := cmd.CombinedOutput()
	if out == nil && err == nil {
	}

	err2 := os.Chmod("script.sh", 0755)
	if err2 != nil {
		log.Fatalf("cmd.Run() failed with \n", err2)
		return false
	}
	return true
}
