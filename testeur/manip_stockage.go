package testeur

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

/**
retourne le numéro  et le Nom du dernier défi enregistré
*/

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
func Clear(path string, exception []string) bool {
	dirRead, _ := os.Open(path)
	dirFiles, _ := dirRead.Readdir(0)
	var excp bool
	fmt.Println(exception)
	// Loop over the directory's files.
	for index := range dirFiles {
		excp = false
		fileHere := dirFiles[index]

		// Get name of file and its full path.
		nameHere := fileHere.Name()
		fullPath := path + nameHere
		for _, name := range exception {
			if name == nameHere {
				excp = true
			}
		}
		// Remove the file.
		if !excp {
			os.Remove(fullPath)
		}
	}
	return true
}

/**
Fonction qui retourne un tableau contenant tous les noms des fichiers du répertoire entré en paramètre
*/
func GetFiles(path string) []string {

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
	for _, file := range GetFiles(path) {
		if file == fileName {
			return true
		}
	}
	return false
}

/*
Renome un fichier "name" par un nouveau Nom "newName" qui se trouve dans le repertoire "pathFile"
*/
func Rename(pathFile string, name string, newName string) {
	name = pathFile + name
	newName = pathFile + newName
	if _, err := exec.Command("sudo", "mv", name, newName).CombinedOutput(); err != nil {
		fmt.Println("error Rename\n", err)
	}
}

// https://golangcode.com/write-data-to-a-csv-file/
func CreateCSV(file_name string, num int) {
	ResultatCSV := BDD.GetParticipant(num)

	file, _ := os.Create(file_name)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Login", "Nom", "Prenom", "Mail", "Num défi", "Résultat", "Nb tentative"})
	for _, value := range ResultatCSV {
		line := []string{value.Resultat.Login,
			value.Etudiant.Nom,
			value.Etudiant.Prenom,
			value.Etudiant.Mail,
			strconv.Itoa(value.Resultat.Defi),
			strconv.Itoa(value.Resultat.Etat),
			strconv.Itoa(value.Resultat.Tentative)}

		if err := writer.Write(line); err != nil {
			fmt.Println(err.Error())
		}
	}
}

//testé
func GetConfigTest(path string) JeuDeTest {
	var Jeu JeuDeTest
	var testUnique CasTest
	var arg Retour
	f, err := os.Open(path + "config")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("os.Open = ", f)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		testUnique.Nom = scanner.Text()
		for _, args := range strings.Split(scanner.Text(), " ") {
			arg.Nom = args
			arg.Contenu = contenu(path + args) // changer le path
			testUnique.arguments = append(testUnique.arguments, arg)
		}

		Jeu.CasDeTest = append(Jeu.CasDeTest, testUnique)
	}
	return Jeu
}

func contenu(path string) string {
	//retourne soit le contenu du fichier, soit l'arborescence
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return "pas de fichier"
	}
	fmt.Println(f)
	fileStat, _ := f.Stat()
	if fileStat.IsDir() {
		output, err := exec.Command("tree", "-A", path).CombinedOutput()
		if err != nil {
			fmt.Println(err.Error())
		}
		return string(output)
	} else {
		output, err := exec.Command("cat", path).CombinedOutput()
		if err != nil {
			fmt.Println(err.Error())
		}
		return string(output)
	}
	return ""
}
