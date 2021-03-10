"use strict";

document.addEventListener('DOMContentLoaded', function() { // Au chargement de la page

    // Instanciation des "date pickers" de Materialize
    var elems = document.querySelectorAll('.datepicker');
    let currentDate = new Date();
    let currentYear = currentDate.getFullYear();
    let nextYear = currentYear + 1;
    let optionsDatePicker = {
        format : "yyyy-mm-dd",
        minDate: new Date(currentYear + "-02-15"),
        maxDate: new Date(nextYear + "-06-30"),
        defaultDate : currentDate
    }
    var instances = M.Datepicker.init(elems, optionsDatePicker);

    // Instanciation des inputs "select" de Materialize
    var selectList = document.querySelectorAll('select');
    var instancesSelect = M.FormSelect.init(selectList);

    // Instanciation des onglets de Materialize
    let el = document.querySelector(".tabs");
    let optionsTabs = {
        duration : 500
    }
    var instanceTab = M.Tabs.init(el, optionsTabs);
    // Selectionne le dernier onglet selectionné
    instanceTab.select(localStorage.getItem("current-tab"));

    // Change les labels des chose-file de défi et de test
    changeInputFile("#file-defi");
    changeInputFile("#file-test");

});

// Stock dans le localstorage le dernier onglet selectionner, pour se replacer dessus au rechargement de la page
let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => li.addEventListener('click', function() {
    let href = li.getAttribute("href").substring(1);
    localStorage.setItem("current-tab", href);
}));


// -------------- Fonctions --------------

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

/*
Function qui permet de modifier un label par le fichier entré dans un input
 */
function changeInputFile(selector) {
    let input = document.querySelector(selector + " input");
    let label = document.querySelector(selector + " label");
    input.addEventListener("change", function() {
        label.innerHTML = getFileName(input.value);
    });
}

/*
Function qui permet de modifier les valeur par défaut dans des input date, suivant le défi selectionné
 */
function ChangeDateInput(event, divID) {
    fetch("http://localhost:8192/GetDefis")
        .then(response => response.json())
        .then(data => {
            let defiActuel = data.find(el => el.Num == event.target.value);
            console.log(defiActuel)
            let datepicker = document.querySelectorAll('div#'+divID+' input.datepicker')
            datepicker[0].value = defiActuel.Date_debut;
            datepicker[1].value = defiActuel.Date_fin;
        })
        .catch(err => console.log(err))
}

function checkJeuDeTestSent(event) {
    fetch("/GetDefis")
        .then(response => response.json())
        .then(data =>  {
            let defiActuel = data.find(el => el.Num == event.target.value);
            console.log(defiActuel)

            let para = document.querySelector('#TestDeposer');
            if (defiActuel.Jeu_de_test) {
                para.innerHTML = "Vous avez déjà déposé un jeu de test pour ce défi."
            } else {
                para.innerHTML = "Vous n'avez pas encore déposé de jeu de test pour ce défi."
            }
        })
}

async function init() {
    const response = await fetch('/GetDefiActuel');
    // waits until the request completes...
    let defiActuel = response.json()
    let para = document.querySelector('#TestDeposer');
    if (defiActuel.Jeu_de_test) {
        para.innerHTML = "Vous avez déjà déposé un jeu de test pour ce défi."
    } else {
        para.innerHTML = "Vous n'avez pas encore déposé de jeu de test pour ce défi."
    }
}


init()