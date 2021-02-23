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
    console.log(localStorage.getItem("current-tab"));
    instanceTab.select(localStorage.getItem("current-tab"));

});

let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => {
    li.addEventListener('click', () => {
        let tab = document.querySelector(".active");
        localStorage.setItem("current-tab", tab.getAttribute("href").substring(1));
        console.log("stock " + tab.getAttribute("href").substring(1));
    });
});

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

function ChangeDateInput(event) {
    console.log("value: " + event.target.value)
}
