document.addEventListener('DOMContentLoaded', function() {
    let el = document.querySelector(".tabs");
    let optionsTabs = {
        duration: 600
    }
    var instanceTab = M.Tabs.init(el, optionsTabs);
    instanceTab.select(localStorage.getItem("current-tab"));

});

let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => li.addEventListener('click', function() {
    let href = li.getAttribute("href").substring(1);
    localStorage.setItem("current-tab", href);
    console.log("h");
}));