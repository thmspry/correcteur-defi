document.addEventListener('DOMContentLoaded', function() {
    var elems = document.querySelectorAll('.datepicker');
    let optionsDatePicker = {
        format : "yyyy-mm-dd",
        month :
            [
                'Janvier',
                'February',
                'March',
                'April',
                'May',
                'June',
                'July',
                'August',
                'September',
                'October',
                'November',
                'December'
            ]
    }
    var instances = M.Datepicker.init(elems, optionsDatePicker);
});

let el = document.querySelector(".tabs");
let optionsTabs = {
    duration : 500
}
var instance = M.Tabs.init(el, optionsTabs);

