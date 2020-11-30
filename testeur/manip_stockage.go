package testeur

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Defi_actuel() (string, string) {
	defis, err := exec.Command("find", path_defis, "-type", "f").CombinedOutput()
	if err != nil {
		fmt.Print("error : ", err)
	}
	liste_defis := strings.Split(string(defis), "\n")
	//Récupere le dernier défis de la liste_defis, split par / et récupere seulement le nom du défi
	nom_defi := strings.Split(liste_defis[len(liste_defis)-2], "/")[4]

	return string([]rune(nom_defi)[5]), nom_defi
}

func clear(path string) bool {
	dirRead, _ := os.Open(path)
	dirFiles, _ := dirRead.Readdir(0)

	// Loop over the directory's files.
	for index := range dirFiles {
		fileHere := dirFiles[index]

		// Get name of file and its full path.
		nameHere := fileHere.Name()
		fullPath := path + nameHere

		// Remove the file.
		os.Remove(fullPath)
		fmt.Println("Removed file:", fullPath)
	}
	return true
}

// fonction qui déplace un fichier (en ayant précisé son chemin pour le trouver) dans un nouveau dossier
func deplacer(file string, path_out string) bool {
	if _, err := exec.Command("mv", file, path_out).CombinedOutput(); err != nil {
		fmt.Println(file, " not found")
		return false
	}
	return true
}

//testé
func getArbo(path string) string {

	out, _ := exec.Command("find", path, "-type", "f").CombinedOutput()
	return string(out)
}
