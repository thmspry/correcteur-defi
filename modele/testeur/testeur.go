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

/*
Fonction a appelé pour tester
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

	os.Rename(modele.PathDefis+correction, PathDirTest+correction)
	os.Rename(modele.PathScripts+scriptEtu, PathDirTest+scriptEtu)
	os.Rename(modele.PathJeuDeTests+jeuDeTest, PathDirTest+jeuDeTest)

	os.Chmod(PathDirTest+scriptEtu, 0700) // script_E197051L_1 exécutable uniquement par l'utilisateur qui le possède
	os.Chmod(PathDirTest+correction, 0700)
	exec.Command("chmod", "-R", "555", PathDirTest+jeuDeTest).Run() //5 = r-x
	// r pour que les scripts puissent lire le contenu des cas de test
	// x pour qu'il puisse accéder/entrer dans le dossier de cas de test

	tabTest, err := getConfigTest(PathDirTest+jeuDeTest, jeuDeTest)
	if err != nil {
		etatTestGlobal = 0
		logs.WriteLog("testeur", "pas de fichier config dans le dossier "+PathDirTest+jeuDeTest)
		messageDeRetour = "Pas de fichier de config"
		resTest = nil
	}

	for i := 0; i < len(tabTest); i++ {

		res := testeurUnique(correction, scriptEtu, login, tabTest[i], PathDirTest)
		res.CasTest = tabTest[i]
		resTest = append(resTest, res)

		manipStockage.Clear(PathDirTest, []string{jeuDeTest, correction, scriptEtu})

		if res.Etat == 0 {
			DAO.SaveResultat(login, numDefi, 0, resTest, false)
			messageDeRetour = "Le cas de test n°" + strconv.Itoa(i) + " n'est pas passé"
			etatTestGlobal = 0
			i = len(tabTest)
		}
		if res.Etat == -1 {
			DAO.SaveResultat(login, numDefi, 0, resTest, false)
			messageDeRetour = "Il y a eu un erreur lors du CasTest n°" + strconv.Itoa(i)
			etatTestGlobal = 0
			i = len(tabTest)
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
- execute le script_E197051L_1 du defis
- stock le résultat de celui-ci
- regarde si il y a eu des nouveaux fichiers qui ont été crée ou non
Si oui :
	- stock le contenu des nouveaux fichiers et leurs Nom dans une map
	- lance le script_E197051L_1 etu
	- stock le contenu des fichiers et leurs noms dans une map étudiant
	- compare le contenu des fichiers
	- return 1 si c'est pareil, 0 sinon
Si non :
	- lance le script_E197051L_1 étu
	- stock son retour imédiat (stdout)
	- compare stdout du défis et de l'étudiant
	- retunr 1 si c'est pareil, 0 sinon
*/
func testeurUnique(correction string, script_user string, login string, test modele.CasTest, PathDirTest string) modele.ResultatTest {
	//o, _ := exec.Command("ls", "-R", "-l", PathDirTest).CombinedOutput()
	//fmt.Println(string(o))
	var stdoutEtu, stdoutCorrection, stderr bytes.Buffer
	args := make([]string, 0)
	for _, arg := range test.Arguments {
		args = append(args, arg.Nom)
	}
	argsString := strings.Join(args, " ")
	res := modele.ResultatTest{
		Etat:           0,
		Res_etu:        make([]modele.Retour, 0),
		Res_correction: make([]modele.Retour, 0),
		Error_message:  "",
	}
	retour := modele.Retour{
		Nom:     "",
		Contenu: "",
	}
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
	arboApres := manipStockage.GetFiles(PathDirTest)
	if len(arboAvant) != len(arboApres) {
		//modif dans un new fichier
		//trouver le fichier / Nom du fichier modifié
		diff := difference(arboAvant, arboApres)
		mapCorrection := make(map[string]string)
		mapEtu := make(map[string]string)
		//boucle qui enregistre les fichiers créé par le script_E197051L_1 de correction
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
		//execution script_E197051L_1 étudiant

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

func TestArtificielReussite(login string) (string, []modele.ResultatTest) {

	var resTest []modele.ResultatTest
	var res modele.ResultatTest

	retour1 := modele.Retour{
		Nom:     "test1",
		Contenu: "plop",
	}
	retour2 := modele.Retour{
		Nom:     "test1",
		Contenu: "plop",
	}
	Castest := modele.CasTest{
		Nom:       "1",
		Arguments: []modele.Retour{retour1, retour2},
	}
	res = modele.ResultatTest{
		Etat:           1,
		CasTest:        Castest,
		Res_etu:        []modele.Retour{retour1, retour2},
		Res_correction: []modele.Retour{retour1, retour2},
		Error_message:  "",
	}

	resTest = append(resTest, res)

	res = modele.ResultatTest{
		Etat:           1,
		CasTest:        Castest,
		Res_etu:        []modele.Retour{retour1, retour2},
		Res_correction: []modele.Retour{retour1, retour2},
		Error_message:  "",
	}

	resTest = append(resTest, res)

	return "Vous avez passé tous les tests avec succès", resTest
}

func TestArtificielEchec(login string) (string, []modele.ResultatTest) {

	var resTest []modele.ResultatTest
	var res modele.ResultatTest

	retour1 := modele.Retour{
		Nom:     "test1",
		Contenu: "plop",
	}
	retour2 := modele.Retour{
		Nom:     "test1",
		Contenu: "plop",
	}
	Castest := modele.CasTest{
		Nom:       "1",
		Arguments: []modele.Retour{retour1, retour2},
	}
	res = modele.ResultatTest{
		Etat:           1,
		CasTest:        Castest,
		Res_etu:        []modele.Retour{retour1, retour2},
		Res_correction: []modele.Retour{retour1, retour2},
		Error_message:  "",
	}

	resTest = append(resTest, res)

	res = modele.ResultatTest{
		Etat:           0,
		CasTest:        Castest,
		Res_etu:        []modele.Retour{retour1, retour2},
		Res_correction: []modele.Retour{retour1, retour2},
		Error_message:  "Il y a eu une erreur quelque part",
	}

	resTest = append(resTest, res)

	res = modele.ResultatTest{
		Etat:           0,
		CasTest:        Castest,
		Res_etu:        []modele.Retour{retour1, retour2},
		Res_correction: []modele.Retour{retour1, retour2},
		Error_message:  "Il y a eu une erreur quelque part",
	}

	resTest = append(resTest, res)

	return "Un ou plusieurs cas de test ont échoués", resTest
}

//testé

func getConfigTest(path string, jt string) ([]modele.CasTest, error) {
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
				arg.Nom = jt + argument
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
