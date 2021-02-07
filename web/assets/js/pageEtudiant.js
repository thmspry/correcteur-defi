"use strict";

document.addEventListener('DOMContentLoaded', function() {

    let input = document.querySelector("input#input-file");
    let label = document.querySelector("#file label");
    //let image = document.querySelector("#file .chose-file img");
    input.addEventListener("change", function() {
        label.innerHTML = label.innerHTML.replace("Choisir un fichier", input.value);
        //image.remove();
    });

});