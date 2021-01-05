package testeur

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

/**
retourne le numéro  et le nom du dernier défi enregistré
*/
func Defi_actuel() (int, string) {
	var files []string
	fileInfo, err := ioutil.ReadDir(Path_defis)
	if err != nil {
		fmt.Print(err)
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return len(files) - 1, files[len(files)-1]
}

func Nb_test(path string) int {
	var files []string
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Print(err)
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return len(files)
}

/*
Fonction qui delete tous les fichiers d'un répertoire
*/
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

/**
Fonction qui retourne un tableau contenant tous les noms des fichiers du répertoire entré en paramètre
*/
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

func Contains(path string, fileName string) bool {
	for _, file := range getFiles(path) {
		if file == fileName {
			return true
		}
	}
	return false
}

/*
Renome un fichier "name" par un nouveau nom "newName" qui se trouve dans le repertoire "pathFile"
*/
func rename(pathFile string, name string, newName string) {
	name = pathFile + name
	newName = pathFile + newName
	if _, err := exec.Command("sudo", "mv", name, newName).CombinedOutput(); err != nil {
		fmt.Println("error rename\n", err)
	}
}
