"use strict";
/**
 * Lance un toast contenant le message placé dans la balise invisible d'ID "alertMsg"
 */
document.addEventListener('DOMContentLoaded', function() {
    let alertMsg = document.querySelector('#alertMsg').innerHTML
    let duree;
    if (alertMsg.length < 10) {
        duree = 3000; // 3 secondes
    } else {
        duree = alertMsg.length * 220; // calcul du temps d'affichage suivant la durée du texte (220 ms par caractère)
    }

    let options = {
        html: alertMsg,
        displayLength : duree,
        classes: 'rounded'
    }
    console.log(alertMsg);
    M.toast(options);
});