package testeur

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Resultat struct {
	Etat           int
	CasTest        CasTest
	Res_etu        []Retour
	Res_correction []Retour
	Error_message  string
}
type Retour struct { // changer le nom --> dossier/fichier
	Nom     string
	Contenu string
}

type JeuDeTest struct {
	CasDeTest []CasTest
}

type CasTest struct {
	nom       string
	arguments []Retour
}

/*
Fonction a appelé pour tester
*/
func Test(login string) (string, []Resultat) {

	var resTest = make([]Resultat, 0) // Resultat
	messageDeRetour := ""
	etatTestGlobal := 0 // État du test des tests (1,0,-1)

	OS := runtime.GOOS
	if OS == "windows" || OS == "darwin" {
		return "Le testeur ne peut être lancé que sur linux", nil
	}

	//Création du user
	if err := exec.Command("useradd", login).Run(); err != nil {
		fmt.Println("error create user : ", err)
	}

	//Associe le dossier à l'user
	if err := exec.Command("mkhomedir_helper", login).Run(); err != nil {
		fmt.Println("error create dir : ", err)
	}
	//Associe le chemin de CasTest au dossier créé
	config.Path_dir_test = "/home/" + login + "/"
	if err := exec.Command("chmod", "770", config.Path_dir_test).Run(); err != nil {
		fmt.Println("error chmod "+config.Path_dir_test, err)
	}
	Clear(config.Path_dir_test, nil)

	//Récupérer le défi actuel
	numDefi := BDD.GetDefiActuel().Num
	//numDefi = BDD.Defi_actuel()
	correction := "correction_" + strconv.Itoa(numDefi)
	scriptEtu := "script_" + login + "_" + strconv.Itoa(numDefi) + ".sh"
	jeuDeTest := "test_defi_" + strconv.Itoa(numDefi) + "/"

	// Début du testUnique

	// Donne les droits d'accès et de modifications à l'étudiant
	exec.Command("sudo", "chown", login, config.Path_scripts+scriptEtu).Run()
	exec.Command("sudo", "chown", login, config.Path_dir_test).Run()

	os.Rename(config.Path_defis+correction, config.Path_dir_test+correction)
	os.Rename(config.Path_scripts+scriptEtu, config.Path_dir_test+scriptEtu)
	os.Rename(config.Path_jeu_de_tests+jeuDeTest, config.Path_dir_test+jeuDeTest)
	exec.Command("chmod", "-R", "444", config.Path_dir_test+jeuDeTest)

	var configTest JeuDeTest
	if Contains(config.Path_jeu_de_tests+jeuDeTest, "config") {
		configTest = GetConfigTest(config.Path_dir_test + jeuDeTest)
		fmt.Println(configTest)
	} else {
		fmt.Println("pas de fichier de config")
		return "pas de fichier de config", nil
	}

	for i := 0; i < len(configTest.CasDeTest); i++ {

		os.Rename(config.Path_dir_test+jeuDeTest, config.Path_dir_test+"test")

		res := testeurUnique(correction, scriptEtu, login, configTest.CasDeTest[i])
		res.CasTest = configTest.CasDeTest[i]
		resTest = append(resTest, res)

		os.Rename(config.Path_dir_test+"test", config.Path_dir_test+jeuDeTest)

		Clear(config.Path_dir_test, []string{jeuDeTest, correction, scriptEtu})

		if res.Etat == 0 {
			BDD.SaveResultat(login, numDefi, 0, false)
			messageDeRetour = "Le CasTest n°" + strconv.Itoa(i) + " n'est pas passé"
			etatTestGlobal = 0
			i = len(configTest.CasDeTest)
		}
		if res.Etat == -1 {
			BDD.SaveResultat(login, numDefi, 0, false)
			messageDeRetour = "Il y a eu un erreur lors du CasTest n°" + strconv.Itoa(i)
			etatTestGlobal = -1
			i = len(configTest.CasDeTest)
		}
	}
	os.Rename(config.Path_dir_test+correction, config.Path_defis+correction)
	os.Rename(config.Path_dir_test+scriptEtu, config.Path_defis+scriptEtu)
	os.Rename(config.Path_dir_test+jeuDeTest, config.Path_defis+jeuDeTest)

	//supprime l'user et son dossier
	if err := exec.Command("sudo", "userdel", login).Run(); err != nil {
		fmt.Println("error sudo userdel : ", err)
	}
	if err := exec.Command("sudo", "rm", "-rf", "/home/"+login).Run(); err != nil {
		fmt.Println("error sudo rm -rf /home/EXXX : ", err)
	}

	if etatTestGlobal == 0 || etatTestGlobal == -1 {
		return messageDeRetour, resTest
	}
	BDD.SaveResultat(login, numDefi, etatTestGlobal, false)
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
func testeurUnique(correction string, script_user string, login string, test CasTest) Resultat {
	args := make([]string, 0)
	for _, arg := range test.arguments {
		args = append(args, arg.Nom)
	}
	argsString := strings.Join(args, " ")
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

	cmd := exec.Command(config.Path_dir_test+correction, argsString)
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
		//boucle qui enregistre les fichiers créé par le script de correction
		for _, name := range diff {
			//On donne seulement le droit de lecture sur les jeux de CasTest
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

		command := "'" + config.Path_dir_test + script_user + " " + argsString + "'"
		cmd := exec.Command("sudo", "-H", "-u", login, "bash", "-c", command)
		cmd.Dir = config.Path_dir_test
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("erreur execution script étudiant : ", err)
			res.Error_message = "erreur execution du script de l'étudiant"
			res.Etat = -1
			return res
		}
		//boucle qui enregistre les fichiers créé par le script de l'étudiant
		for _, name := range diff {
			f, err := exec.Command("cat", config.Path_dir_test+name).CombinedOutput()
			if err != nil {
				fmt.Println("erreur execution cat : ", config.Path_dir_test+name, "\n", err)
				res.Error_message = "erreur lecture du fichier " + config.Path_dir_test + name
				res.Etat = -1
				return res
			}
			mapEtu[name] = string(f)
			retour.Nom = name
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
		cmd := exec.Command("sudo", "-H", "-u", login, "bash", "-c", command)
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
