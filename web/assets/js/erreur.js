"use strict";
document.addEventListener('DOMContentLoaded', function() {
    let msgError = document.querySelector('#MsgError')
    let options = {
        html: msgError.textContent,
        classes: 'rounded'
    }
    console.log(msgError.textContent)
    M.toast(options);
});