package manipStockage

import (
	"bufio"
	"encoding/csv"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
)

/*
@Clear Fonction qui delete tous les fichiers d'un répertoire
@path chemin menant au répertoire
@exception tableau contenant les noms des fichiers que la fonction ne doit pas supprimer
*/
func Clear(path string, exception []string) {
	dirRead, _ := os.Open(path)
	dirFiles, _ := dirRead.Readdir(0)
	var excp bool
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
}

/**
@GetFiles return un tableau contenant tous les noms des fichiers du répertoire entré en paramètre
@path chemin menant au répertoire
*/
func GetFiles(path string) []string {
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

/**
@Contains
@return true si fileName appartient au répertoire path
		false sinon
*/
func Contains(path string, fileName string) bool {
	for _, file := range GetFiles(path) {
		if file == fileName {
			return true
		}
	}
	return false
}

// https://golangcode.com/write-data-to-a-csv-file/
/**
@CreateCSV créer un CSV contenant les résultats des étudiants pour un défi
@fileName nom du fichier
@num numéro du défi dont on veut avoir les résultats
*/
func CreateCSV(fileName string, num int) {
	ResultatCSV := DAO.GetParticipants(num)

	file, _ := os.Create(fileName)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Login", "Nom", "Prenom", "Mail", "Num défi", "Résultat", "Nb tentative"})
	for _, value := range ResultatCSV {
		line := []string{value.Resultat.Login,
			value.Etudiant.Nom,
			value.Etudiant.Prenom,
			value.Etudiant.Mail(),
			strconv.Itoa(value.Resultat.Defi),
			strconv.Itoa(value.Resultat.Etat),
			strconv.Itoa(value.Resultat.Tentative)}

		if err := writer.Write(line); err != nil {
			logs.WriteLog("func CreateCSV", "Erreur écriture de la ligne")
		}
	}
}

/**
@Contenu fonction utilisé par getConfig
@path chemin vers le fichier/dossier
@return le contenu de path si c'est un fichier,
sinon retourne l'arborescence de path si c'est un dossier
*/
func Contenu(path string) string {
	f, err := os.Open(path)
	if err != nil {
		logs.WriteLog("fonction contenu", "Erreur fichier "+path+" introuvable")
		return "pas de fichier"
	}
	fileStat, _ := f.Stat()
	if fileStat.IsDir() {
		output, _ := exec.Command("tree", "-A", path).CombinedOutput()
		return string(output)
	} else {
		output, _ := exec.Command("cat", path).CombinedOutput()
		return string(output)
	}
}

/**
@GetTriche regroupe les étudiants ayant rendu un script similaire (minimun 2 étudiants par groupe)
@numDefi numéro du défi concerné
@return un tableau contenant les groupes, chaque groupe est un tableau contenant la liste des logins des étudiants ayant un script similaire
*/
func GetTriche(numDefi int) [][]string {
	participants := DAO.GetParticipants(numDefi)

	type script struct {
		Login   string
		Contenu string
	}
	var s script
	TabParticipants := make([]script, 0)

	for _, part := range participants {
		s.Login = part.Etudiant.Login
		f, _ := ioutil.ReadFile(modele.PathScripts + "script_" + s.Login + "_" + strconv.Itoa(numDefi))
		s.Contenu = string(f)
		TabParticipants = append(TabParticipants, s)
	}
	TabSimilaire := make([][]script, 0)
	for _, s = range TabParticipants {
		b := true
		for i, sim := range TabSimilaire {
			if s.Contenu == sim[0].Contenu {
				sim = append(sim, s)
				TabSimilaire[i] = sim
				b = false
			}
		}
		if b {
			TabSimilaire = append(TabSimilaire, []script{s})
		}

	}
	res := make([][]string, 0)
	for _, groupe := range TabSimilaire {
		g := make([]string, 0)
		for _, tricheur := range groupe {
			g = append(g, tricheur.Login)
		}
		if len(g) >= 2 {
			res = append(res, g)
		}
	}
	return res
}

/**
@GetFileLineByLine permet de récupérer chaque ligne du fichier sous la forme d'un tableau
@path chemin vers le fichier
@return un tableau contenant le contenu du fichier path
*/
func GetFileLineByLine(path string) []string {
	f, err := os.Open(path)
	tab := make([]string, 0)
	if err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			tab = append(tab, scanner.Text())
		}
	}
	return tab
}
