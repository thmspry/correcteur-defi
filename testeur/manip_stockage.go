package testeur

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Defi_actuel() (string, string) {
	defis, err := exec.Command("find", path_defis, "-type", "f").CombinedOutput()
	if err != nil {
		fmt.Print("error : ", err)
	}
	liste_defis := strings.Split(string(defis), "\n.")
	num := strconv.Itoa(len(liste_defis) - 1)
	return num, "defi_" + num + ".sh"
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
		//fmt.Println("Removed file:", fullPath)
	}
	return true
}

// fonction qui déplace un fichier (en ayant précisé son chemin pour le trouver) dans un nouveau dossier
func deplacer(file string, path_out string) bool {
	fmt.Println("deplacer :" + file + " vers " + path_out)
	if _, err := exec.Command("sudo", "mv", file, path_out).CombinedOutput(); err != nil {
		fmt.Println(file, " not found\n", err.Error())
		return false
	}
	return true
}

//testé
func getFiles(path string) []string {

	//out, _ := exec.Command("find", path, "-type", "f").CombinedOutput()
	//return string(out)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	} else {
		t := len(files)
		var Files = make([]string, t)
		for i := 0; i < t; i++ {
			Files[i] = files[i].Name()
		}
		return Files
	}

	return nil
}

// fonction qui rend le fichier executable
func makeFileExecutable(script string) bool {
	if err := os.Chmod(script, 0755); err != nil {
		fmt.Print("chmod on ", script, " failed")
		return false
	}
	return true
}

func rename(pathFile string, name string, newName string) {
	name = pathFile + name
	newName = pathFile + newName
	if _, err := exec.Command("sudo", "mv", name, newName).CombinedOutput(); err != nil {
		fmt.Println("error rename\n", err)
	}
}
