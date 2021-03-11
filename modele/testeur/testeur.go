package testeur

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/config"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
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
type Retour struct { // changer le Nom --> dossier/fichier
	Nom     string
	Contenu string
}

type JeuDeTest struct {
	CasDeTest []CasTest
}

type CasTest struct {
	Nom       string
	Arguments []Retour
}

/*
Fonction a appelé pour tester
*/
func Test(login string) (string, []Resultat) {

	var resTest = make([]Resultat, 0) // Resultat
	messageDeRetour := ""
	etatTestGlobal := 1 // État du test des tests (1,0,-1)

	OS := runtime.GOOS
	if OS == "windows" || OS == "darwin" {
		return "Le testeur ne peut être lancé que sur linux", nil
	}

	//Création du user
	if err := exec.Command("useradd", login).Run(); err != nil {
		logs.WriteLog("testeur", "Erreur création de l'utilisateur "+login)
		return "Erreur création de l'utilisateur", nil
	}

	//Associe le dossier à l'user
	if err := exec.Command("mkhomedir_helper", login).Run(); err != nil {
		logs.WriteLog("testeur", "Erreur création dossier de l'utilisateur "+login)
		return "Erreur création du dossier de l'utilisateur", nil
	}
	//Associe le chemin de CasTest au dossier créé
	Path_dir_test := "/home/" + login + "/"
	if err := exec.Command("chmod", "770", Path_dir_test).Run(); err != nil {
		logs.WriteLog("testeur", "Erreur chmod du dossier "+Path_dir_test)
	}
	manipStockage.Clear(Path_dir_test, nil)

	//Récupérer le défi actuel
	numDefi := BDD.GetDefiActuel().Num
	//numDefi = BDD.DefiActuel()
	correction := "correction_" + strconv.Itoa(numDefi)
	scriptEtu := "script_" + login + "_" + strconv.Itoa(numDefi)
	jeuDeTest := "test_defi_" + strconv.Itoa(numDefi) + "/"

	// Début du testUnique

	// Donne les droits d'accès et de modifications à l'étudiant
	exec.Command("sudo", "chown", login, config.Path_scripts+scriptEtu).Run()
	exec.Command("sudo", "chown", login, Path_dir_test).Run()

	os.Rename(config.Path_defis+correction, Path_dir_test+correction)
	os.Rename(config.Path_scripts+scriptEtu, Path_dir_test+scriptEtu)
	os.Rename(config.Path_jeu_de_tests+jeuDeTest, Path_dir_test+jeuDeTest)

	os.Chmod(Path_dir_test+scriptEtu, 0700) // script exécutable uniquement par l'utilisateur qui le possède
	os.Chmod(Path_dir_test+correction, 0700)
	exec.Command("chmod", "-R", "555", Path_dir_test+jeuDeTest).Run() //5 = r-x
	// r pour que les scripts puissent lire le contenu des cas de test
	// x pour qu'il puisse accéder/entrer dans le dossier de cas de test

	var configTest JeuDeTest
	if manipStockage.Contains(Path_dir_test+jeuDeTest, "config") {
		configTest = getConfigTest(Path_dir_test+jeuDeTest, jeuDeTest)
	} else {
		logs.WriteLog("testeur", "pas de fichier config dans le dossier "+Path_dir_test+jeuDeTest)
		return "Pas de fichier de config", nil
	}

	fmt.Println("configtest : \n", configTest)

	for i := 0; i < len(configTest.CasDeTest); i++ {

		res := testeurUnique(correction, scriptEtu, login, configTest.CasDeTest[i], Path_dir_test)
		res.CasTest = configTest.CasDeTest[i]
		resTest = append(resTest, res)

		manipStockage.Clear(Path_dir_test, []string{jeuDeTest, correction, scriptEtu})

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
	os.Rename(Path_dir_test+correction, config.Path_defis+correction)
	os.Rename(Path_dir_test+scriptEtu, config.Path_scripts+scriptEtu)
	os.Rename(Path_dir_test+jeuDeTest, config.Path_jeu_de_tests+jeuDeTest)

	//supprime l'user et son dossier
	exec.Command("sudo", "userdel", login).Run()
	exec.Command("sudo", "rm", "-rf", "/home/"+login).Run()

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
	- stock le contenu des nouveaux fichiers et leurs Nom dans une map
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
func testeurUnique(correction string, script_user string, login string, test CasTest, PathDirTest string) Resultat {
	//o, _ := exec.Command("ls", "-R", "-l", PathDirTest).CombinedOutput()
	//fmt.Println(string(o))
	var stdout, stderr bytes.Buffer
	args := make([]string, 0)
	for _, arg := range test.Arguments {
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
	arboAvant := manipStockage.GetFiles(PathDirTest)

	cmd := exec.Command(PathDirTest+correction, argsString)
	cmd.Dir = PathDirTest
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		logs.WriteLog("testeur ("+login+")", "Erreur execution script de correction : "+string(stderr.Bytes()))
		res.Error_message = "erreur execution du script de correction\n" + string(stderr.Bytes())
		res.Etat = -1
		return res
	}
	stdout_correction := string(stdout.Bytes())
	arboApres := manipStockage.GetFiles(PathDirTest)
	if len(arboAvant) != len(arboApres) {
		//modif dans un new fichier
		//trouver le fichier / Nom du fichier modifié
		diff := difference(arboAvant, arboApres)
		mapCorrection := make(map[string]string)
		mapEtu := make(map[string]string)
		//boucle qui enregistre les fichiers créé par le script de correction
		for _, name := range diff {
			//On donne seulement le droit de lecture sur les jeux de CasTest
			exec.Command("chmod", "777", PathDirTest+name).Run()
			f, err := exec.Command("cat", PathDirTest+name).CombinedOutput()
			if err != nil {
				logs.WriteLog("testeur ("+login+")", "Erreur lecture du fichier "+PathDirTest+name)
				res.Error_message = "erreur lecture du fichier " + PathDirTest + name + " (correction)"
				res.Etat = -1
				return res
			}
			mapCorrection[name] = string(f)
			retour.Nom = "fichier " + name
			retour.Contenu = string(f)

			res.Res_correction = append(res.Res_correction, retour)
		}
		//execution script étudiant

		cmdUser := PathDirTest + script_user + " " + argsString
		cmd := exec.Command("sudo", "-H", "-u", login, "bash", "-c", cmdUser)
		cmd.Dir = PathDirTest
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			logs.WriteLog("testeur ("+login+")", "Erreur execution script etudiant "+login+": "+string(stderr.Bytes()))
			res.Error_message = "erreur execution du script de l'étudiant\n" + string(stderr.Bytes())
			res.Etat = -1
			return res
		}

		//boucle qui enregistre les fichiers créé par le script de l'étudiant
		for _, name := range diff {
			f, err := exec.Command("cat", PathDirTest+name).CombinedOutput()
			if err != nil {
				logs.WriteLog("testeur ("+login+")", "Erreur lecture du fichier "+PathDirTest+name)
				res.Error_message = "erreur lecture du fichier " + PathDirTest + name + " (etudiant)"
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
		retour.Nom = "sortie standart"
		retour.Contenu = string(stdout_correction)
		res.Res_correction = append(res.Res_correction, retour)
		cmdUser := PathDirTest + script_user + " " + argsString
		cmd := exec.Command("sudo", "-H", "-u", login, "bash", "-c", cmdUser)
		cmd.Dir = PathDirTest
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		err := cmd.Run()
		if err != nil {
			logs.WriteLog("testeur ("+login+")", "Erreur execution du script étudiant : "+string(stderr.Bytes()))
			res.Error_message = "erreur execution du script étudiant\n" + string(stderr.Bytes())
			res.Etat = -1
			return res
		}
		stdout_etu := string(stdout.Bytes())
		retour.Contenu = string(stdout_etu)
		res.Res_etu = append(res.Res_etu, retour)
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

func TestArtificiel(login string) (string, []Resultat) {

	var resTests = make([]Resultat, 0) // Resultats
	res := Resultat{
		Etat:           0,
		Res_etu:        make([]Retour, 0),
		Res_correction: make([]Retour, 0),
		Error_message:  "",
	}

	var retoursEtu = make([]Retour, 0) // Retours
	ret11 := Retour{
		Nom:     "grp1",
		Contenu: "François\nPatrice\nDaniel",
	}
	ret21 := Retour{
		Nom:     "grp2",
		Contenu: "Paul\nThomas\nGabriel",
	}
	retoursEtu = append(retoursEtu, ret11)
	retoursEtu = append(retoursEtu, ret21)
	res.Res_etu = retoursEtu

	var retoursCor = make([]Retour, 0) // Retours
	ret12 := Retour{
		Nom:     "grp1",
		Contenu: "François\nPatrice\nPaul",
	}
	ret22 := Retour{
		Nom:     "grp2",
		Contenu: "Daniel\nThomas\nGabriel",
	}
	retoursCor = append(retoursCor, ret12)
	retoursCor = append(retoursCor, ret22)
	res.Res_correction = retoursCor

	resTests = append(resTests, res)

	retCasTest := Retour{
		Nom:     "Retour cas de test",
		Contenu: "le contenu du retour du cas test",
	}

	var retoursCasTest = make([]Retour, 0)
	retoursCasTest = append(retoursCasTest, retCasTest)
	casTest := CasTest{
		Nom:       "Le nom du cas de test",
		Arguments: retoursCasTest,
	}

	res.CasTest = casTest

	return "Vous avez passé tous les tests avec succès", resTests
}

//testé
func getConfigTest(path string, jt string) JeuDeTest {
	var Jeu JeuDeTest
	var testUnique CasTest
	f, _ := os.Open(path + "config")
	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		testUnique.Nom = strconv.Itoa(i)
		for _, args := range strings.Split(scanner.Text(), " ") {
			testUnique.Arguments = append(testUnique.Arguments, Retour{
				jt + args,
				manipStockage.Contenu(path + args),
			})
		}
		Jeu.CasDeTest = append(Jeu.CasDeTest, testUnique)
		testUnique.Arguments = nil
		i++
	}
	return Jeu
}
