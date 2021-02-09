"use strict";

document.addEventListener('DOMContentLoaded', function() {

    let input = document.querySelector("input#input-file");
    let label = document.querySelector("#file label");
    let ligneCommande = document.querySelector("#ligne-commande");
    const idEtu = document.querySelector("#id-etu").innerHTML;
    const debutLigne = document.createElement("span");
    debutLigne.setAttribute("class", "green-text");
    debutLigne.innerHTML = idEtu + "@iut";
    ligneCommande.append(debutLigne);
    ligneCommande.append("$     ");
    input.addEventListener("change", function() {
        let fileName = getFileName(input.value);
        ligneCommande.append("./" + fileName);
        label.innerHTML = label.innerHTML.replace("Choisir un fichier", fileName);
    });

});

function getFileName(filePath) {
    let filePathSplit;
    if (filePath.includes("/")) {   // Linux
        filePathSplit = filePath.split("/");
    } else {                                    // Windows
        filePathSplit = filePath.split("\\");
    }
    return filePathSplit[filePathSplit.length-1];
}
