"use strict";


/**
 * Fonction assyncrone pour récuperer tous les défis en Json
 * @returns {Promise<*>}
 */
async function getDefis() {
    let response = await fetch("/GetDefis");
    let data = await response.json();
    data = JSON.stringify(data);
    data = JSON.parse(data);
    return data;
}

document.addEventListener('DOMContentLoaded', function() { // Au chargement de la page

    getDefis().then(data => {
        // Initialisation du sélecteur de dates Materialiaze
        var elems = document.querySelectorAll('.datepicker');
        var optionsDatePicker = {}

        if (data!=null) { // S'il y a un/des défi(s)
            console.log(data)
            // La date minimale qu'on peut choisir pour un date est la date de fin du dernier défi le plus récent
            let maxDate = data[data.length - 1].DateFin;
            console.log("La max date est alors : "+ maxDate)
            let currentDate = new Date();
            let currentYear = currentDate.getFullYear();
            let nextYear = currentYear + 1;
            optionsDatePicker = {
                format: "yyyy-mm-dd",
                minDate: new Date(maxDate),
                maxDate: new Date(nextYear + "-06-30"),
                defaultDate: currentDate
            }
        } else { // S'il n'y a pas de défi
            let currentDate = new Date();
            let currentYear = currentDate.getFullYear();
            let nextYear = currentYear + 1;
            optionsDatePicker = {
                format: "yyyy-mm-dd",
                minDate: new Date(currentYear + "-02-15"),
                maxDate: new Date(nextYear + "-06-30"),
                defaultDate: currentDate
            }

        }
        var instancesDate = M.Datepicker.init(elems, optionsDatePicker);
        var timers = document.querySelectorAll('.timepicker')
        var instancesTime = M.Timepicker.init(timers, {
            twelveHour:false
        });

    })



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
    changeInputFile("#file-defi-modify");

});

// Stock dans le localstorage le dernier onglet sélectionné, pour se replacer dessus au rechargement de la page
let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => li.addEventListener('click', function() {
    let href = li.getAttribute("href").substring(1);
    localStorage.setItem("current-tab", href);
}));


// -------------- Fonctions --------------

/*
Fonction qui permet de récupérer seulement le nom du fichier à partir d'un path en paramètre
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
    fetch("/GetDefis")
        .then(response => response.json())
        .then(data => {
            let defiSelect = data.find(el => el.Num == event.target.value);
            console.log(defiSelect)
            let dateDParse = defiSelect.DateDebut.split('T')
            let dateFParse = defiSelect.DateFin.split('T')
            let datepicker = document.querySelectorAll(`div#${divID} input.datepicker`)
            datepicker[0].value = dateDParse[0];
            datepicker[1].value = dateFParse[0];
            let timepicker = document.querySelectorAll(`div#${divID} input.timepicker`)
            timepicker[0].value = dateDParse[1].slice(0,5)
            timepicker[1].value = dateFParse[1].slice(0,5)
        })
        .catch(err => console.log(err))
}

function checkJeuDeTestSent(event) {
    fetch("/GetDefis")
        .then(response => response.json())
        .then(data =>  {
            let defiSelect = data.find(el => el.Num == event.target.value);

            let para = document.querySelector('#TestDeposer');
            if (defiSelect.JeuDeTest) {
                para.innerHTML = "Vous avez déjà déposé un jeu de test pour ce défi."
            } else {
                para.innerHTML = "Vous n'avez pas encore déposé de jeu de test pour ce défi."
            }
        })
}

async function init() {
    const defiActuel = await fetch('/GetDefiActuel')
        .then((response) => response.json())
        .then((data) => {
        return data
    });
    // waits until the request completes...
    let dateDParse = defiActuel.DateDebut.split('T')
    let dateFParse = defiActuel.DateFin.split('T')
    let para = document.querySelector('#TestDeposer');
    if (defiActuel.JeuDeTest) {
        para.innerHTML = "Vous avez déjà déposé un jeu de test pour ce défi."
    } else {
        para.innerHTML = "Vous n'avez pas encore déposé de jeu de test pour ce défi."
    }
}


init()