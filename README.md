#Application Web - Correction des défis du lundi
_Initié par Loïg Jezequel._ \
_Développé par Paul Vernin, Matteo Ordrenneau, Thomas Peray et Antoine Gru._

GitLab : https://gitlab.univ-nantes.fr/E192543L/projet-s3/

Cette application web vise à simplifier le travail nécessaire à la réalisation des défis du lundi proposé aux premières années de l'IUT Informatique de Nantes, dans le module Introduction aux Systèmes Informatiques.
Elle utilisable d'une part par les élèves, et d'une autre part par les enseignants.

##1. Installation
Il est nécessaire d'installer le projet dans un environnement pouvant exécuter du Goland, ainsi que gérer une base de données sous SQLite3.

Une fois l'archive téléchargée, placez-vous à la racine.
3 possibilité d'execution sont disponibles :
 - "init" : initialise l'application
 - "reset" : supprime les fichiers des dossiers et la base de données
 - "start" : lance l'application


Pour une première utilisation, il sera nécessaire de réaliser les 3. Exécutez donc ces 3 commandes :
> go run main.go init \
> go run main.go reset \
> go run main.go start

Il permettra notamment d'installer les packages externes à Goland. \
Par la suite, les exécution future nécessiteront uniquement la commande `go run main.go start`.
Pour arrêter l'exécution de l'application, le raccourci Ctrl + C suffit.

Pour générer un nouvel exécutable, saisissez dans le terminal :
> set GOOS=linux \
> set GOARCH=amd6 \
> go build

##2. Utilisation
- **Pour une utilisation en local** : \
Suite au lancement (`go run main.go start`), connectez vous à l'adresse : http://localhost:8192/login \
L'application s'execute sur le port 8192


- **Pour une utilisation sur serveur** : \
  +TODO \
  Pour initialiser le projet, il est nécessaire d'ajouter les fichiers .html et .css
  sur le serveur à la suite.


Par défaut, il existe un compte Admin : `login : admin | password : admin`




