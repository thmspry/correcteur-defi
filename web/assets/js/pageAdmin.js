document.addEventListener('DOMContentLoaded', function() {
    var elems = document.querySelectorAll('.datepicker');
    let optionsDatePicker = {
        format : "yyyy-mm-dd",
        minDate: new Date("2021-02-15"),
        maxDate: new Date("2022-06-30"),
        defaultDate : new Date()
    }
    var instances = M.Datepicker.init(elems, optionsDatePicker);


    changeInputFile("#file-defi");
    changeInputFile("#file-test");

    var selectList = document.querySelectorAll('select');
    var instancesSelect = M.FormSelect.init(selectList);

    var instanceTab = M.Tabs.init(el, optionsTabs);
    instanceTab.select(localStorage.getItem("current-tab"));

});

let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => li.addEventListener('click', function() {
    let href = li.getAttribute("href").substring(1);
    localStorage.setItem("current-tab", href);
}));


let el = document.querySelector(".tabs");
let optionsTabs = {
    duration : 500
}
var instance = M.Tabs.init(el, optionsTabs);

function getFileName(filePath) {
    let filePathSplit;
    if (filePath.includes("/")) {   // Linux
        filePathSplit = filePath.split("/");
    } else {                                    // Windows
        filePathSplit = filePath.split("\\");
    }
    return filePathSplit[filePathSplit.length-1];
}

function changeInputFile(selector) {
    let input = document.querySelector(selector + " input");
    let label = document.querySelector(selector + " label");
    input.addEventListener("change", function() {
        label.innerHTML = getFileName(input.value);
    });
}

document.addEventListener('DOMContentLoaded', function() {
    var elems = document.querySelectorAll('select');
    var instances = M.FormSelect.init(elems);

});

function ChangeDateInput(event, divID) {
    fetch("http://localhost:8192/GetDefis")
        .then(response => response.json())
        .then(data => {
            defiActuel = data.find(el => el.Num == event.target.value);
            datepicker = document.querySelectorAll('div#'+divID+' input.datepicker')
            datepicker[0].value = defiActuel.Date_debut;
            datepicker[1].value = defiActuel.Date_fin;

        })
        .catch(err => console.log(err))
}
