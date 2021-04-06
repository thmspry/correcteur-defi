document.addEventListener('DOMContentLoaded', function() {
    let el = document.querySelector(".tabs");
    let optionsTabs = {
        duration: 600
    }
    var instanceTab = M.Tabs.init(el, optionsTabs);
    instanceTab.select(localStorage.getItem("current-tab"));

    // Pour renvoyer sur l'onglet login après l'inscription
    let registerButton = document.querySelector("#register .button");
    registerButton.addEventListener("click", function () {
        localStorage.setItem("current-tab", "login");
    })

});

let tabLi = document.querySelectorAll(".tabs a")
tabLi.forEach(li => li.addEventListener('click', function() {
    let href = li.getAttribute("href").substring(1);
    localStorage.setItem("current-tab", href);
}));