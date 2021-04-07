"use strict";
document.addEventListener('DOMContentLoaded', function() {
    let alertMsg = document.querySelector('#alertMsg')
    let options = {
        html: alertMsg.textContent,
        classes: 'rounded'
    }
    console.log(msgError.textContent)
    M.toast(options);
});