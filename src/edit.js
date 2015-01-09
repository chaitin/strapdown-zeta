//add ace 
var editor = ace.edit("editor");
editor.setTheme("ace/theme/monokai");
editor.getSession().setMode("ace/mode/markdown");
editor.getSession().setTabSize(2);
editor.getSession().setUseSoftTabs(true);

//bind event
var sav = document.getElementById("savValue");
var form = document.getElementsByTagName('form')[0];
form.addEventListener("submit",function(){
    sav.value = editor.getValue();
});

