//add ace 
var editor = ace.edit("editor");
editor.setTheme("ace/theme/twilight");
editor.getSession().setMode("ace/mode/markdown");

//bind event
var sav = document.getElementById("savValue");
var form = document.getElementsByTagName('form')[0];
form.addEventListener("submit",function(){
    sav.value = editor.getValue();
});

