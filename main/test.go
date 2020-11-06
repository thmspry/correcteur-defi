package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"projS3/main/testeur"
)

func main() {
	fmt.Println("testeur retourne : ", testeur.Test("EXXX"))
	/*db,_ := sql.Open("sqlite3", "./projS3.db")
	//stmt,_ := db.Prepare("CREATE TABLE etudiant (id INTEGER PRIMARY KEY, firstname TEXT)")
	//stmt.Exec()
	stmt, _ := db.Prepare("INSERT INTO etudiant (firstname) VALUES (?)")
	rows, _ := db.Query("SELECT id, firstname FROM etudiant")
	stmt.Exec("ouisqd")
	var id int
	var firstname string
	for rows.Next() {
		rows.Scan(&id, &firstname)
		fmt.Printf(strconv.Itoa(id) + ": "+ firstname + "\n")
	}*/
}
