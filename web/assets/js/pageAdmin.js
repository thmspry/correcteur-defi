"use strict";


/**
 * Fonction asyncrone pour récupérer tous les défis en Json
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
            // La date minimale qu'on peut choisir pour un date est la date de fin du dernier défi le plus récent
            let maxDate = data[data.length - 1].DateFin;
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
            twelveHour:false,
            defaultTime: new Date().toLocaleTimeString()
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

    // Chargement des graphiques
    fetch("/GetParticipantsDefis")
        .then(response => response.json())
        .then(data => {
            let participants = data.Participants;
            // Graphique general
            google.charts.load('current', {'packages':['corechart']});
            google.charts.setOnLoadCallback(drawChart1);
            function drawChart1() {
                let tab = [['defis', 'participations\n', 'Reussite\n']]
                participants.forEach(item => {
                    tab.push(['defi ' + item.Num, item.ParticipantsDefi, item.Reussite])
                })
                let data = google.visualization.arrayToDataTable(tab);

                let options = {
                    title: 'Evolution de la participation et du taux de reussite au cours des défis',
                    curveType: 'function',
                    legend: { position: 'bottom' },
                    vAxis: {
                        viewWindowMode:'explicit',
                        viewWindow: {
                            min: 0
                        }
                    }
                };

                let chart = new google.visualization.LineChart(document.getElementById('curve_chart'));

                chart.draw(data, options);
            }
        })
        .catch(err => console.log(err))

    // Graphique camembert
    google.charts.load('current', {'packages':['corechart']});
    google.charts.setOnLoadCallback(drawChart);

    function drawChart() {
        const data = google.visualization.arrayToDataTable([
            ['Participants', 'participation'],
            ['participants', 0]
        ]);
        const options = {
            title: 'Selectionner un défi'
        };
        const chart = new google.visualization.PieChart(document.getElementById('piechart'));
        chart.draw(data, options);
    }

    // graphique  Nombre moyen de tentatives
    google.charts.load('current', {packages: ['corechart', 'bar']});
    google.charts.setOnLoadCallback(drawStacked);

    function drawStacked() {

        const data = google.visualization.arrayToDataTable([
            ['defi', 'valeur'],
            ['defi', 0],
        ]);

        const options = {
            title: 'Selectionner un défi',
            hAxis: {
                minValue: 0,
                maxValue:100
            },
        };
        const chart = new google.visualization.BarChart(document.getElementById('chart_div'));
        chart.draw(data, options);
    }




});

// Stock dans le localstorage le dernier onglet sélectionné, pour se replacer dessus au rechargement de la page
let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => li.addEventListener('click', function() {
    let href = li.getAttribute("href").substring(1);
    localStorage.setItem("current-tab", href);
}));


// -------------- Fonctions --------------



/**
 * Fonction qui permet de récupérer seulement le nom du fichier à partir d'un path en paramètre
 * @param filePath path du fichier comportant des '/' ou '\'
 * @returns {*} le nom du fichier
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

/**
 * Function qui permet de modifier un label par le fichier entré dans un input
 * @param selector l'input voulu
 */
function changeInputFile(selector) {
    let input = document.querySelector(selector + " input");
    let label = document.querySelector(selector + " label");
    input.addEventListener("change", function() {
        label.innerHTML = getFileName(input.value);
    });
}


/**
 * Function qui permet de modifier les valeur par défaut dans des input date, suivant le défi selectionné
 * @param event l'événement
 * @param divID l'ID de la div comportant l'input Date
 * @constructor
 */
function ChangeDateInput(event, divID) {
    fetch("/GetDefis")
        .then(response => response.json())
        .then(data => {
            let defiSelect = data.find(el => el.Num == event.target.value);
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

/**
* Function qui permet de modifier les graphiques des statistiques
* @param event l'événement
* @constructor
*/
function ChangeDefisStats(event) {
    if(document.querySelector('#selectStatsDefi').options[document.querySelector('#selectStatsDefi').selectedIndex].value !== "") {
        fetch("/GetParticipantsDefis")
            .then(response => response.json())
            .then(data => {
                const select = document.querySelector('#selectStatsDefi');
                const nbDefi = parseInt(select.options[select.selectedIndex].value);
                let participants = data.Participants.filter(elem => elem.Num === nbDefi);
                participants = participants[0];
                const nonParticipants = data.NbEtudiants - participants.ParticipantsDefi;
                const data1 = google.visualization.arrayToDataTable([
                    ['Defi', 'Taux de participation'],
                    ['Participants', participants.ParticipantsDefi],
                    ['Non Participants', nonParticipants],
                ]);

                const options1 = {
                    title: 'Taux de participation defi n°' + nbDefi
                };

                const chart1 = new google.visualization.PieChart(document.getElementById('piechart'));

                chart1.draw(data1, options1);

                const data2 = google.visualization.arrayToDataTable([
                    ['defi', 'valeur'],
                    ['defi', participants.MoyenneTentatives],
                ]);

                const options2 = {
                    title: 'Nombre moyen de tentatives',
                    chartArea: {width: '50%'},
                    hAxis: {
                        title: 'Nombre moyen de tentatives',
                        minValue: 0,
                        maxValue:100
                    },
                };

                const chart2 = new google.visualization.BarChart(document.getElementById('chart_div'));
                chart2.draw(data2, options2);
            })
            .catch(err => console.log(err))
    }
}

/**
 * Vérifie si le jeu de test à été envoyé et modifier l'HTML en conséquence
 * @param event l'événement
 */
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

/**
 * Mets en place les date au chargement de la page
 * @returns {Promise<void>} la promesse
 */
async function init() {
    const ListeDefis = await fetch('/GetDefis')
        .then((response) => response.json())
        .then((data) => {
        return data
    });
    // waits until the request completes...
    let para = document.querySelector('#TestDeposer');
    if (ListeDefis[0].JeuDeTest) {
        para.innerHTML = "Vous avez déjà déposé un jeu de test pour ce défi."
    } else {
        para.innerHTML = "Vous n'avez pas encore déposé de jeu de test pour ce défi."
    }
}

// Appelé à chaque (re)chargement
init()