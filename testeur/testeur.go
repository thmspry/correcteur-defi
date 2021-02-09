package testeur

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strconv"
)

type Resultat struct {
	Etat           int
	Test           string
	Res_etu        []Retour
	Res_correction []Retour
	Error_message  string
}
type Retour struct {
	Nom     string
	Contenu string
}

type JeuDeTest struct {
	CasDeTest []CasTest
}

type CasTest struct {
	nom       string
	arguments []string
}

/*
Fonction a appelé pour tester
*/
func Test(etudiant string) (string, []Resultat) {

	messageDeRetour := ""
	etatTestGlobal := 0

	OS := runtime.GOOS
	if OS == "windows" || OS == "darwin" {
		return "Le testeur ne peut être lancé que sur linux", nil
	}

	//Création du user
	if err := exec.Command("useradd", etudiant).Run(); err != nil {
		fmt.Println("error create user : ", err)
	}

	//Associe le dossier à l'user
	if err := exec.Command("mkhomedir_helper", etudiant).Run(); err != nil {
		fmt.Println("error create dir : ", err)
	}
	//Associe le chemin de Test au dossier créé
	config.Path_dir_test = "/home/" + etudiant + "/"
	if err := exec.Command("chmod", "770", config.Path_dir_test).Run(); err != nil {
		fmt.Println("error chmod "+config.Path_dir_test, err)
	}
	clear(config.Path_dir_test)

	//Récupérer le défi actuel
	numDefi := BDD.GetDefiActuel().Num
	//numDefi = BDD.Defi_actuel()
	correction := "correction_" + strconv.Itoa(numDefi)
	scriptEtu := "script_" + etudiant + "_" + strconv.Itoa(numDefi) + ".sh"
	PathJeuDeTestDuDefi := config.Path_jeu_de_tests + "test_defi_" + strconv.Itoa(numDefi) + "/"
	var resTest = make([]Resultat, 0)

	// Début du testUnique

	// Donne les droits d'accès et de modifications à l'étudiant
	exec.Command("sudo", "chown", etudiant, config.Path_scripts+scriptEtu).Run()
	exec.Command("sudo", "chown", etudiant, config.Path_dir_test).Run()

	var configTest JeuDeTest
	if Contains(PathJeuDeTestDuDefi, "config") {
		configTest = GetConfigTest(PathJeuDeTestDuDefi)
		fmt.Println(configTest)
	} else {
		fmt.Println("pas de fichier de config")
		return "pas de fichier de config", nil
	}

	for i := 0; i < len(configTest.CasDeTest); i++ {
		deplacer(config.Path_defis+correction, config.Path_dir_test)
		deplacer(config.Path_scripts+scriptEtu, config.Path_dir_test)
		test := "test_" + strconv.Itoa(i)
		deplacer(PathJeuDeTestDuDefi+test, config.Path_dir_test)
		exec.Command("chmod", "444", config.Path_dir_test+test).Run()
		rename(config.Path_dir_test, test, "Test")

		res := testeurUnique(correction, scriptEtu, etudiant)
		f, _ := ioutil.ReadFile(config.Path_dir_test + test)
		res.Test = string(f)
		resTest = append(resTest, res)

		rename(config.Path_dir_test, "Test", test)
		deplacer(config.Path_dir_test+test, PathJeuDeTestDuDefi)
		deplacer(config.Path_dir_test+correction, config.Path_defis)
		deplacer(config.Path_dir_test+scriptEtu, config.Path_scripts)
		clear(config.Path_dir_test)

		if res.Etat == 0 {
			BDD.SaveResultat(etudiant, numDefi, 0, false)
			messageDeRetour = "Le Test n°" + strconv.Itoa(i) + " n'est pas passé"
			etatTestGlobal = 0
			break
		}
		if res.Etat == -1 {
			BDD.SaveResultat(etudiant, numDefi, 0, false)
			messageDeRetour = "Il y a eu un erreur lors du Test n°" + strconv.Itoa(i)
			etatTestGlobal = -1
			break
		}
	}

	clear(config.Path_dir_test)

	//supprime l'user et son dossier
	if err := exec.Command("sudo", "userdel", etudiant).Run(); err != nil {
		fmt.Println("error sudo userdel : ", err)
	}
	if err := exec.Command("sudo", "rm", "-rf", "/home/"+etudiant).Run(); err != nil {
		fmt.Println("error sudo rm -rf /home/EXXX : ", err)
	}

	BDD.SaveResultat(etudiant, numDefi, 1, false)

	if etatTestGlobal == 0 || etatTestGlobal == -1 {
		return messageDeRetour, resTest
	}
	return "Vous avez passé tous les tests avec succès", resTest
}

/*
Fonctionnement du testeur unique :
- execute le script du defis
- stock le résultat de celui-ci
- regarde si il y a eu des nouveaux fichiers qui ont été crée ou non
Si oui :
	- stock le contenu des nouveaux fichiers et leurs nom dans une map
	- lance le script etu
	- stock le contenu des fichiers et leurs noms dans une map étudiant
	- compare le contenu des fichiers
	- return 1 si c'est pareil, 0 sinon
Si non :
	- lance le script étu
	- stock son retour imédiat (stdout)
	- compare stdout du défis et de l'étudiant
	- retunr 1 si c'est pareil, 0 sinon
*/
func testeurUnique(correction string, script_user string, etudiant string) Resultat {

	res := Resultat{
		Etat:           0,
		Res_etu:        make([]Retour, 0),
		Res_correction: make([]Retour, 0),
		Error_message:  "",
	}
	retour := Retour{
		Nom:     "",
		Contenu: "",
	}
	arboAvant := GetFiles(config.Path_dir_test)

	cmd := exec.Command(config.Path_dir_test + correction)
	cmd.Dir = config.Path_dir_test
	stdout_correction, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("erreur execution script défis : ", err)
		res.Error_message = "erreur execution du script de correction"
		res.Etat = -1
		return res
	}
	arboApres := GetFiles(config.Path_dir_test)
	if len(arboAvant) != len(arboApres) {
		//modif dans un new fichier
		//trouver le fichier / nom du fichier modifié
		diff := difference(arboAvant, arboApres)
		mapCorrection := make(map[string]string)
		mapEtu := make(map[string]string)
		for _, name := range diff {
			//On donne seulement le droit de lecture sur les jeux de Test
			exec.Command("chmod", "777", config.Path_dir_test+name).Run()
			f, err := exec.Command("cat", config.Path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", config.Path_dir_test+name, "\n", err)
				res.Error_message = "erreur lecture du fichier " + config.Path_dir_test + name
				res.Etat = -1
				return res
			}
			mapCorrection[name] = string(f)
			retour.Nom = "fichier " + name
			retour.Contenu = string(f)

			res.Res_correction = append(res.Res_correction, retour)
		}
		//execution script étudiant

		command := "'" + config.Path_dir_test + script_user + "'"
		cmd := exec.Command("sudo", "-H", "-u", etudiant, "bash", "-c", command)
		cmd.Dir = config.Path_dir_test
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("erreur execution script étudiant : ", err)
			res.Error_message = "erreur execution du script de l'étudiant"
			res.Etat = -1
			return res
		}
		//Récup les fichiers
		for _, name := range diff {
			f, err := exec.Command("cat", config.Path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", config.Path_dir_test+name, "\n", err)
				res.Error_message = "erreur lecture du fichier " + config.Path_dir_test + name
				res.Etat = -1
				return res
			}
			mapEtu[name] = string(f)
			retour.Nom = "fichier " + name
			retour.Contenu = string(f)
			res.Res_etu = append(res.Res_etu, retour)
			if mapEtu[name] != mapCorrection[name] {
				res.Etat = 0
				return res
			}
		}
		res.Etat = 1
		return res
	} else {
		command := "'" + config.Path_dir_test + script_user + "'"
		cmd := exec.Command("sudo", "-H", "-u", etudiant, "bash", "-c", command)
		cmd.Dir = config.Path_dir_test
		stdout_etu, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err, stdout_etu)
			res.Error_message = "erreur execution du script étudiant"
			res.Etat = -1
			return res
		}

		retour.Nom = "sortie standart"
		retour.Contenu = string(stdout_etu)
		res.Res_etu = append(res.Res_etu, retour)
		retour.Contenu = string(stdout_correction)
		res.Res_correction = append(res.Res_correction, retour)
		if res.Res_etu[0] == res.Res_correction[0] {
			res.Etat = 1
			return res
		} else {
			res.Etat = 0
			return res
		}
	}
}

/*
Fonction qui renvoie un tableau contenant la différence entre les deux tableaux entrés en paramètre
*/
func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return diff
}
