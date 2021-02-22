document.addEventListener('DOMContentLoaded', function() {
    var elems = document.querySelectorAll('.datepicker');
    let optionsDatePicker = {
        format : "yyyy-mm-dd",
        minDate: new Date("2021-02-15"),
        maxDate: new Date("2022-06-30")
    }
    var instances = M.Datepicker.init(elems, optionsDatePicker);



    changeInputFile("#file-defi");
    changeInputFile("#file-test");


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

