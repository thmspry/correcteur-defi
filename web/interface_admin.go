package web

import (
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/BDD"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/testeur"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Admin struct {
}

type data_pageAdmin struct {
	Etudiants []BDD.Etudiant
	Defis_etu []BDD.Defi
}

func pageAdmin(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		data := data_pageAdmin{
			Etudiants: BDD.GetEtudiants(),
			Defis_etu: nil,
		}

		if r.URL.Query()["Etudiant"] != nil {
			etu := r.URL.Query()["Etudiant"][0]
			data.Defis_etu = BDD.GetDefis(etu)
		}

		t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))

		if err := t.Execute(w, data); err != nil {
			log.Printf("error exec template : ", err)
		}
	}
	if r.Method == "POST" {

		r.ParseMultipartForm(10 << 20)

		file, _, err := r.FormFile("defi")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()

		num, _ := testeur.Defi_actuel()
		path := ""
		submit := r.FormValue("submit")
		if submit == "defi_upload" {
			n, _ := strconv.Atoi(num)
			num = strconv.Itoa(n + 1)
			path = "./ressource/defis/defi_" + num + ".sh"
		} else if submit == "modification" {
			path = "./ressource/defis/defi_" + num + ".sh"
		} else if submit == "test_upload" {
			path = "./ressource/jeu_de_test/test_defi_" + num + "/"
			num_test := testeur.Nb_test(path)
			path = "./ressource/jeu_de_test/test_defi_" + num + "/test_" + strconv.Itoa(num_test)
		}

		script, err := os.Create(path) // remplacer handler.Filename par le nom et on le place où on veut

		if err != nil {
			fmt.Println("Internal Error")
			fmt.Println(err)
		}

		defer script.Close()

		_, err = io.Copy(script, file)
		if err != nil {
			fmt.Println("Internal Error")
			fmt.Println(err)
			return
		}

		os.Chmod(path, 770)

		// return that we have successfully uploaded our file!
		fmt.Println("Successfully Uploaded File\n")
		//rename fonctionne pas jsp pk
		//os.Rename(handler.Filename, "script_E1000.sh")
		http.Redirect(w, r, "/pageAdmin", http.StatusFound)
	}

}
