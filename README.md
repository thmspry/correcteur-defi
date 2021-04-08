<a href="https://iutnantes.univ-nantes.fr/"><img align="right" src="https://iutnantes.univ-nantes.fr/medias/photo/logoiutq_1377690591795.gif?ID_FICHE=627306" width="100"></a>
<br><br>
<h1>Go-Testeur - Correction des défis du lundi</h1>

*Initié par Loïg Jezequel.*\
*Développé par Paul Vernin, Matteo Ordrenneau, Thomas Peray et Antoine Gru.*

GitLab : https://gitlab.univ-nantes.fr/E192543L/projet-s3/

Cette application web vise à simplifier le travail nécessaire à la réalisation des défis du lundi proposé aux premières années de l'IUT Informatique de Nantes, dans le module Introduction aux Systèmes Informatiques.
Elle utilisable d'une part par les élèves, et d'une autre part par les enseignants.

<h2>1. Installation </h2>
Il est nécessaire d'installer le projet dans un environnement ayant Go (1.15.2 recommandé), ainsi que GCC (9.2 recommandé) pour gérer la database SQLite3.

Une fois l'archive téléchargée, placez-vous à la racine.
3 possibilité d'execution sont disponibles :
 - "init" : initialise l'application
 - "reset" : supprimer les fichiers de ressource et remet la database a l'état initial
 - "start" : lance l'application

Pour une première utilisation, il sera nécessaire de réaliser les 3. Exécutez donc ces 3 commandes :
> go run main.go init <br>
> go run main.go start

<p>Il permettra notamment d'installer les packages externes à Goland. <br>
Par la suite, les exécution future nécessiteront uniquement la commande `go run main.go start`.
Pour arrêter l'exécution de l'application, le raccourci Ctrl + C suffit.<p>

Pour générer un nouvel exécutable à mettre sur le serveur, saisissez dans le terminal :
> set GOOS=linux && set GOARCH=amd64 (sur windows)\
> export GOOS=linux && export GOARCH=amd64 (sur mac)\
> go build

<h2>2. Lancer l'application sur le serveur</h2>
<h4>1) Se connecter au serveur de l'IUT avec Pulse Secure</h4>
Pour la suite il est nécessaire de possèder un utilisateur enregistré sur le serveur.
<h4>2) Importer le projet</h4>

A l'aide de la commande scp, on va importer sur le serveur :
 - l'exécutable généré par go build
 - le fichier mailConf.json
 - le répertoire web

En faisant `scp file user@172.26.82.23:` (ajouter `-R` pour le répertoire web)


<h4>3) Lancer l'application en continue</h4>

Se connecter au serveur avec `ssh user@172.26.82.23` dans un terminal\
Initialiser l'application avec `go run main.go init`\
Ouvrir un screen en faisant `screen `, dans celui-ci exécuter l'application en mode **root** en faisant `sudo ./projet-s3 start`\
Fermer le screen en faisant `CTRL + D + A` et vous pouvez fermer le terminal sans que cela arrête  l'application

<h2>3. Utilisation</h2>

- **Pour une utilisation en local** : \
Suite au lancement (`go run main.go start`), connectez vous à l'adresse : http://localhost:8192 \
L'application s'execute sur le port 8192

- **Pour une utilisation sur serveur** : \
  Il faut en premier lieu se connecter au serveur de l'IUT soit en y étant soit avec Pulse Secure\
  En suite on peut accéder à l'application par navigateur web à l'adresse : http://172.26.82.23:8192

Par défaut lorsqu'on initialise l'application, un compte admin (`login : admin | password : admin`) est créé afin de se connecter à l'interface admin du site



