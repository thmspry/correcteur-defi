"use strict";

document.addEventListener('DOMContentLoaded', function() { // Au chargement de la page

    /* --- Ligne pour modifier la page en conséquence du fichier selectionné dans le input file --- */
    let input = document.querySelector("input#input-file");// On récupère l'input
    if(input != null) { // Lorsqu'il y a un defi en cours (pour éviter les erreur dans la console)
        let label = document.querySelector("#file label");              // On récupère le label
        let ligneCommande = document.querySelector("#ligne-commande");  // On récupère la ligne dans le .res
        const idEtu = document.querySelector("#id-etu").innerHTML;      // On récupère l'ID de l'étudiant présent dans la page
        const debutLigne = document.querySelector(".green-text");       // La ligne de commande est séparée en deux texts
        debutLigne.innerHTML = idEtu + "@iut";                                  // On ajoute a cette partie un text
        ligneCommande.append("$     ");                                         // Puis dans l'autre partie (en blanc)
        input.addEventListener("change", function() {               // Lorsqu'on ajoute un fichier dans le input
            let fileName = getFileName(input.value);                            // On récupère le nom du fichier
            ligneCommande.append("./" + fileName);                              // "./" pour simuler un executablle
            label.innerHTML = label.innerHTML.replace("Choisir un fichier", fileName); // On remplace l'ancien text par le nom du fichier
        });
    }
    let el = document.querySelector(".tabs");
    var instance = M.Tabs.init(el, null);

});

/*
Fonction qui permet de récuper seulement le nom du fichier à partir d'un path en paramètre
 */
function getFileName(filePath) {
    let filePathSplit;
    if (filePath.includes("/")) {   // Linux
        filePathSplit = filePath.split("/");
    } else {                                    // Windows
        filePathSplit = filePath.split("\\");
    }
    return filePathSplit[filePathSplit.length-1];
}