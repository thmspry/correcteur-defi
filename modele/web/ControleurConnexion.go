package web

import (
	"crypto/rand"
	"fmt"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele"
	"gitlab.univ-nantes.fr/E192543L/projet-s3/modele/logs"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"time"
)

type dataConnexion struct {
	Alert string
}

/*
@accueil Traite toutes les requêtes effectuées sur la page d'accueil `/login`
*/
func accueil(w http.ResponseWriter, r *http.Request) {
	data := dataConnexion{
		Alert: "",
	}
	fmt.Println("Méthode de accueil Etudiant :", r.Method)
	// Verification s'il le client à un token enregistré
	if tk, err := r.Cookie("token"); err == nil {
		// S'il est stocké dans la BD : qu'il n'est pas expiré
		if DAO.TokenExiste(tk.Value) {
			fmt.Println("Token existe : ", tk.Value)
			// On détermine s'il s'agit d'un token admin ou étudiant afin de faire les bonnes redirections
			role := DAO.TokenRole(tk.Value)
			if role == "etudiants" {
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
			} else if role == "administrateur" {
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			}
			return
		} else {
			logs.WriteLog("Erreur TOKEN", "Le token n'existe pas")
		}
	} else {
		logs.WriteLog("Aucun cookie trouvé", err.Error())
	}
	// Cas ou la requete client utilise la méthode GET
	if r.Method == "GET" {
		if r.URL.String() == "/login" {
			// On charge le fichier html correspondant à la page d'accueil étudiant
			t, err := template.ParseFiles("./web/html/accueil.html")
			if err != nil {
				fmt.Print("Erreur chargement accueil.html")
				logs.WriteLog("Erreur chargement accueil.html : ", err.Error())
			} else {
				err = t.Execute(w, nil)
				if err != nil {
					logs.WriteLog("Erreur d'éxecution de la page accueil.html : ", err.Error())
				}
			}
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		// Cas ou la requete client utilise la méthode POST
	} else if r.Method == "POST" {
		// S'il l'argument de l'url est login
		if r.URL.String() == "/login?login" {
			// request provient du formulaire pour se connecter
			login := r.FormValue("login")
			password := r.FormValue("password")
			// On récupère les arguments passés dans le body de la requête
			fmt.Println("Tentative de connexion avec :", login, " ", password)
			existe := DAO.LoginCorrect(login, password) // on test le couple login/passwordHashé
			if existe {
				logs.WriteLog(login, "Le couple login/password est correct")
				//Création du token correspondant
				token := tokenGenerator()
				temps := 20 * time.Minute                                               // défini le temps d'attente avant que l'on supprime le token de la BDD
				expiration := time.Now().Add(temps)                                     // défini le temps de validité du token
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration} // Création du token sous forme de cookie
				http.SetCookie(w, &cookie)                                              // On envoit le cookie au client
				fmt.Println("(login=", login, ",token=", token)
				DAO.InsertToken(login, token)
				logs.WriteLog(login, "connexion étudiant")
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound) // On redirige vers la page etudiant correspondant
				go DeleteToken(login, temps)                           // On lance une goroutine qui va supprimer ce token de la BDD lorsqu'il ne sera plus valide
				return
			} else {
				logs.WriteLog(login, "Mot de passe incorrecte connexion étudiant")
				// On redirige le client s'il y a eu une erreur de login
				page, err := template.ParseFiles("./web/html/accueil.html")
				if err != nil {
					fmt.Print("erreur chargement accueil.html")
					logs.WriteLog(login+" Erreur de chargement de la page accueil.html : ", err.Error())
				} else {
					// On passe des données à la vue pour faire afficher un message d'erreur
					data.Alert = "Une erreur login/mot de passe est survenue"
					err = page.Execute(w, data)
					if err != nil {
						logs.WriteLog(login+" Erreur d'éxecution de la page accueil.html : ", err.Error())
					}
				}
			}
			// S'il l'argument de l'url est register
		} else if r.URL.String() == "/login?register" {
			// request provient du formulaire pour s'enregistrer
			// Si le login est déjà utilisé
			if DAO.IsLoginUsed(r.FormValue("login")) {
				// On redirige le client s'il y a eu une erreur d'inscription
				page, err := template.ParseFiles("./web/html/accueil.html")
				if err != nil {
					fmt.Print("Erreur chargement accueil.html : ", err)
					logs.WriteLog("Erreur chargement accueil.html : ", err.Error())
				} else {
					loginExist := r.FormValue("login")
					data.Alert = "L'identifiant entré est déjà utilisé"
					fmt.Println("Ce login (", loginExist, ") existe déjà")
					logs.WriteLog("Already Exist : ", "Ce login ("+loginExist+") existe déjà")
					err = page.Execute(w, data)
					if err != nil {
						logs.WriteLog("Erreur d'éxecution de la page accueil.html avec erreur login existe : ", err.Error())
					}
				}
			} else {
				// Si le login n'est pas déjà utilisé
				data.Alert = ""
				etu := modele.Etudiant{
					Login:    r.FormValue("login"),
					Password: r.FormValue("password"),
					Prenom:   r.FormValue("prenom"),
					Nom:      r.FormValue("nom"),
				}
				passwordHashed, err := bcrypt.GenerateFromPassword([]byte(etu.Password), 14) // hashage du mot de passe
				if err == nil {
					etu.Password = string(passwordHashed) // le mot de passe à stocker est hashé
					DAO.Register(etu)                     // ajouter l'etudiant dans la base de données.
					logs.WriteLog(etu.Login, "création du compte : "+etu.Login+":"+etu.Password)
					http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
				} else {
					logs.WriteLog("Erreur lors du hashage du mot de passe : ", err.Error())
					logs.WriteLog(r.FormValue("login"), "échec création du compte")
					http.Redirect(w, r, "/login", http.StatusFound)
				}
			}
		}
	}
}

/*
	Fonction appelé lorsque l'url parsée est : http://localhost:8192/loginAdmin
*/
func connexionAdmin(w http.ResponseWriter, r *http.Request) {
	data := dataConnexion{
		Alert: "",
	}
	fmt.Println("methode de connexion Admin :", r.Method)
	// Verification s'il le client à un token enregistré
	if tk, err := r.Cookie("token"); err == nil {
		if DAO.TokenExiste(tk.Value) {
			fmt.Println("Token existe : ", tk.Value)
			role := DAO.TokenRole(tk.Value)
			fmt.Println(role)
			if role == "etudiants" {
				http.Redirect(w, r, "/pageEtudiant", http.StatusFound)
			} else if role == "administrateur" {
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
			}
			return
		}
	}

	// Cas ou la requete client utilise la méthode GET
	if r.Method == "GET" {
		// On charge le fichier html correspondant à la page de connexion admin
		if r.URL.String() == "/loginAdmin" {
			page, err := template.ParseFiles("./web/html/connexionAdmin.html")
			if err != nil {
				fmt.Print("erreur chargement connexionAdmin.html : ", err)
				logs.WriteLog("Erreur chargement connexionAdmin.html : ", err.Error())
			} else {
				data.Alert = ""
				err = page.Execute(w, data)
				if err != nil {
					fmt.Print("erreur affichage connexionAdmin.html : ", err)
					logs.WriteLog("Erreur chargement connexionAdmin.html : ", err.Error())
				}
			}
		} else {
			http.Redirect(w, r, "/loginAdmin", http.StatusFound)
		}

	} else if r.Method == "POST" {

		if r.URL.String() == "/loginAdmin?login" {
			login := r.FormValue("login")
			password := r.FormValue("password")
			fmt.Println("Tentative de connexion admin avec :", login, " ", password)
			existe := DAO.LoginCorrectAdmin(login, password) // on test le couple login/passwordHashé
			if existe {
				//Création du token
				token := tokenGenerator()
				temps := 20 * time.Minute // défini le temps d'attente
				expiration := time.Now().Add(temps)
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
				http.SetCookie(w, &cookie)
				DAO.InsertToken(login, token)
				logs.WriteLog(login, "connexion admin réussie, création d'un token")
				http.Redirect(w, r, "/pageAdmin", http.StatusFound)
				go DeleteToken(login, temps)
				return
			} else {
				data.Alert = "mot de passe incorrecte connexion admin"
				logs.WriteLog(login, data.Alert)
				page, err := template.ParseFiles("./web/html/connexionAdmin.html")
				if err != nil {
					logs.WriteLog("Erreur du chargement de la page connexionAdmin.html : ", err.Error())
				} else {
					data.Alert = "Une erreur login/mot de passe est survenue"
					err = page.Execute(w, data)
					if err != nil {
						logs.WriteLog("Erreur du chargement de la page connexionAdmin.html : ", err.Error())
					}
				}
			}
		}
	}
}

/*
	Fonction qui permet de supprimer de la BD le token d'une connexion après une durée donnée
*/
func DeleteToken(login string, temps time.Duration) {
	time.Sleep(temps)
	logs.WriteLog(login, "Déconnexion du serveur")
	DAO.DeleteToken(login)
	return
}

/*
	Fonction qui permet de générer un token
	@return une chaîne de caractères de la forme : %x[.., .., .., ..]
*/
func tokenGenerator() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		logs.WriteLog("Generation token", "il y a eu une erreur lors de la génération du token : "+err.Error())
	}
	return fmt.Sprint("%x", b)
}
