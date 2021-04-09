package testeur

import (
	"bufio"
	"bytes"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/manipStockage"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

/**
@Test fonction qui permet de tester le script d'un étudiant donné en paramètre
@login login de l'étudiant qui fait le test
@return un string donnant le résultat global du test
@return un tableau contenant le résultat obtenu pour chaque cas de test
*/
func Test(login string) (string, []modele.ResultatTest) {

	var resTest = make([]modele.ResultatTest, 0) // ResultatTest
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
		exec.Command("sudo", "userdel", login).Run()
		return "Erreur création du dossier de l'utilisateur", nil
	}
	//Associe le chemin de CasTest au dossier créé
	PathDirTest := "/home/" + login + "/"
	if err := exec.Command("chmod", "770", PathDirTest).Run(); err != nil {
		logs.WriteLog("testeur", "Erreur chmod du dossier "+PathDirTest)
	}
	manipStockage.Clear(PathDirTest, nil)

	//Récupérer le défi actuel
	numDefi := DAO.GetDefiActuel().Num
	correction := "correction_" + strconv.Itoa(numDefi)
	scriptEtu := "script_" + login + "_" + strconv.Itoa(numDefi)
	jeuDeTest := "test_defi_" + strconv.Itoa(numDefi) + "/"

	// Début du testUnique

	// Donne les droits d'accès et de modifications à l'étudiant
	exec.Command("sudo", "chown", login, modele.PathScripts+scriptEtu).Run()
	exec.Command("sudo", "chown", login, PathDirTest).Run()

	//permet déplacer les scripts et le dossier de jeu de test dans le dossier de l'utilisateur créé
	os.Rename(modele.PathDefis+correction, PathDirTest+correction)
	os.Rename(modele.PathScripts+scriptEtu, PathDirTest+scriptEtu)
	os.Rename(modele.PathJeuDeTests+jeuDeTest, PathDirTest+jeuDeTest)

	os.Chmod(PathDirTest+scriptEtu, 0700)                           // exécutable uniquement par l'utilisateur qui le possède (nomé par son login)
	os.Chmod(PathDirTest+correction, 0700)                          // pareil sauf que c'est root
	exec.Command("chmod", "-R", "555", PathDirTest+jeuDeTest).Run() //5 = r-x
	// r pour que les scripts puissent lire le contenu des cas de test
	// x pour qu'il puisse accéder/entrer dans le dossier de cas de test

	tabTest, err := getConfigTest(PathDirTest+jeuDeTest, jeuDeTest) //récupère la configuration des cas de test à passer
	if err != nil {
		etatTestGlobal = 0
		logs.WriteLog("testeur", "pas de fichier config dans le dossier "+PathDirTest+jeuDeTest)
		messageDeRetour = "Pas de fichier de config"
		resTest = nil
	}

	//boucle pour chaque cas de test a effectué
	for i := 0; i < len(tabTest); i++ {

		res := testeurUnique(correction, scriptEtu, login, tabTest[i], PathDirTest)
		res.CasTest = tabTest[i]
		resTest = append(resTest, res)

		manipStockage.Clear(PathDirTest, []string{jeuDeTest, correction, scriptEtu}) //on supprime les potentiels fichiers créés par l'exécution

		if res.Etat == 0 { // s'il le test n'est pas passé
			DAO.SaveResultat(login, numDefi, 0, resTest, false)
			messageDeRetour = "Le cas de test n°" + strconv.Itoa(i) + " n'est pas passé"
			etatTestGlobal = 0
			i = len(tabTest) // on arrête la boucle
		}
		if res.Etat == -1 { //s'il n'y a eu une erreur
			DAO.SaveResultat(login, numDefi, 0, resTest, false)
			messageDeRetour = "Il y a eu un erreur lors du CasTest n°" + strconv.Itoa(i)
			etatTestGlobal = -1
			i = len(tabTest) // on arrête la boucle
		}
	}
	os.Rename(PathDirTest+correction, modele.PathDefis+correction)
	os.Rename(PathDirTest+scriptEtu, modele.PathScripts+scriptEtu)
	os.Rename(PathDirTest+jeuDeTest, modele.PathJeuDeTests+jeuDeTest)

	//supprime l'user et son dossier
	exec.Command("sudo", "userdel", login).Run()
	exec.Command("sudo", "rm", "-rf", "/home/"+login).Run()

	if etatTestGlobal == 1 {
		messageDeRetour = "Vous avez passé tous les tests avec succès"
	}
	logs.WriteLog(login, "testeur défi"+strconv.Itoa(numDefi)+", état : "+strconv.Itoa(etatTestGlobal))
	DAO.SaveResultat(login, numDefi, etatTestGlobal, resTest, false)
	return messageDeRetour, resTest
}

/*
Fonctionnement du testeur unique :
 * execute le script_E197051L_1 du defis
 * stock le résultat de celui-ci
 * regarde s'il y a eu des nouveaux fichiers qui ont été crée ou non
 * Si oui :
 *	- stock le contenu des nouveaux fichiers et leurs noms dans une map Correction
 *	- lance le script_LOGIN_X (script étudiant)
 *	- stock le contenu des fichiers et leurs noms dans une map étudiant
 *	- compare le contenu des deux maps
 *	- return 1 si c'est identique, 0 sinon
 * Si non :
 *	- lance le script_E197051L_1 étu
 * 	- stock son retour imédiat (stdout)
 * 	- compare stdout de la correction et du script étudiant
 *	- retur 1 si c'est identique, 0 sinon
*/
func testeurUnique(correction string, script_user string, login string, test modele.CasTest, PathDirTest string) modele.ResultatTest {
	var stdoutEtu, stdoutCorrection, stderr bytes.Buffer
	var retour modele.Retour
	res := modele.ResultatTest{
		Etat:           0,
		Res_etu:        make([]modele.Retour, 0),
		Res_correction: make([]modele.Retour, 0),
		Error_message:  "",
	}

	//regroupe tous les arguments sous la forme d'un unique string à mettre en parametre de l'exécution du script
	args := make([]string, 0)
	for _, arg := range test.Arguments {
		args = append(args, arg.Nom)
	}
	argsString := strings.Join(args, " ")

	//Enregistre l'arborescence du dossier de test avant l'exécution du script de correction
	arboAvant := manipStockage.GetFiles(PathDirTest)

	cmd := exec.Command(PathDirTest+correction, argsString)
	cmd.Dir = PathDirTest
	cmd.Stderr = &stderr
	cmd.Stdout = &stdoutCorrection
	err := cmd.Run()
	if err != nil {
		logs.WriteLog("testeur ("+login+")", "Erreur execution script_E197051L_1 de correction : "+string(stderr.Bytes()))
		res.Error_message = "erreur execution du script_E197051L_1 de correction\n" + string(stderr.Bytes())
		res.Etat = -1
		return res
	}
	arboApres := manipStockage.GetFiles(PathDirTest) //arbo après exécution pour voir si des fichiers ont été créé ou non
	if len(arboAvant) != len(arboApres) {
		diff := difference(arboAvant, arboApres) //récupère un tableau contenant la liste des fichiers créés
		mapCorrection := make(map[string]string)
		mapEtu := make(map[string]string)
		for _, name := range diff { //boucle qui enregistre les fichiers et leur contenu créé par l'exécution du script de correction
			exec.Command("chmod", "777", PathDirTest+name).Run()
			// donne tous les droits afin que l'utilisateur qui va exécuter le script étudiant puisse réécrire par dessus
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

		//exécution du script étudiant par le nouveau utilisateur du dossier de test
		cmdUser := PathDirTest + script_user + " " + argsString
		cmd := exec.Command("sudo", "-H", "-u", login, "bash", "-c", cmdUser)
		cmd.Dir = PathDirTest
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			logs.WriteLog("testeur ("+login+")", "Erreur execution script_E197051L_1 etudiant "+login+": "+string(stderr.Bytes()))
			res.Error_message = "erreur execution du script_E197051L_1 de l'étudiant\n" + string(stderr.Bytes())
			res.Etat = -1
			return res
		}

		//boucle qui enregistre les fichiers créé par le script_E197051L_1 de l'étudiant
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
		retour.Contenu = string(stdoutCorrection.Bytes())
		res.Res_correction = append(res.Res_correction, retour)
		cmdUser := PathDirTest + script_user + " " + argsString
		cmd := exec.Command("sudo", "-H", "-u", login, "bash", "-c", cmdUser)
		cmd.Dir = PathDirTest
		cmd.Stderr = &stderr
		cmd.Stdout = &stdoutEtu
		err := cmd.Run()
		if err != nil {
			logs.WriteLog("testeur ("+login+")", "Erreur execution du script_E197051L_1 étudiant : "+string(stderr.Bytes()))
			res.Error_message = "erreur execution du script_E197051L_1 étudiant\n" + string(stderr.Bytes())
			res.Etat = -1
			return res
		}
		retour.Contenu = string(stdoutEtu.Bytes())
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

/**
@getConfigTest fonction qui récupere la configuration des tests a effectué sur le script étudiant
@path chemin menant au fichier de configuration
@testDefiX nom du dossier de test
@return un tableau contenant la liste des cas de tests a effectuer
*/
func getConfigTest(path string, testDefiX string) ([]modele.CasTest, error) {
	var tabTest []modele.CasTest
	var testUnique modele.CasTest
	var arg modele.Retour
	f, err := os.Open(path + "config")
	if err != nil {
		logs.WriteLog("getConfigTest open config", err.Error())
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		testUnique.Nom = strconv.Itoa(i)
		for _, argument := range strings.Split(scanner.Text(), " ") {
			arg.Contenu = manipStockage.Contenu(path + argument)
			if arg.Contenu == "pas de fichier" {
				arg.Nom = argument
			} else {
				arg.Nom = testDefiX + argument
			}
			testUnique.Arguments = append(testUnique.Arguments, arg)
			arg = modele.Retour{}
		}
		tabTest = append(tabTest, testUnique)
		testUnique.Arguments = nil
		i++
	}
	return tabTest, nil
}
