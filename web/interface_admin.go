package web

import (
	"fmt"
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

func pageAdmin(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t := template.Must(template.ParseFiles("./web/html/pageAdmin.html"))

		if err := t.Execute(w, nil); err != nil {
			log.Printf("error exec template : ", err)
		}
	}
	if r.Method == "POST" {

		r.ParseMultipartForm(10 << 20)

		file, _, err := r.FormFile("script_etu")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()

		num, _ := testeur.Defi_actuel()

		submit := r.FormValue("submit")
		if submit == "upload" {
			n, _ := strconv.Atoi(num)
			num = strconv.Itoa(n - 1)
		} else if submit == "modification" {

		}

		script, err := os.Create("./ressource/defis/defi_" + num + ".sh") // remplacer handler.Filename par le nom et on le place où on veut

		if err != nil {
			fmt.Println("Internal Error")
			fmt.Println(err)
			return
		}

		defer script.Close()

		_, err = io.Copy(script, file)
		if err != nil {
			fmt.Println("Internal Error")
			fmt.Println(err)
			return
		}

		// return that we have successfully uploaded our file!
		fmt.Println("Successfully Uploaded File\n")
		//rename fonctionne pas jsp pk
		//os.Rename(handler.Filename, "script_E1000.sh")
		http.Redirect(w, r, "/pageAdmin", http.StatusFound)
	}

}
