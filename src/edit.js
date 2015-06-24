;(function(window, document){

  var store = new Persist.Store('strapdown_editor', { swf_path: '/persist.swf' }),
  filename = window.location.pathname,
  markdownEl = document.getElementsByTagName('xmp')[0] || document.getElementsByTagName('textarea')[0];

  store.get(filename, function(ok, value){

    //add ace 
    var editor = ace.edit("editor"),
    session = editor.getSession();

    editor.setTheme("ace/theme/monokai");
    session.setMode("ace/mode/markdown");
    session.setTabSize(2);
    session.setUseSoftTabs(true);

    if (ok && value){
      //TODO: Add timestamp check?
      session.setValue(value)
    }
    //bind event
    var sav = document.getElementById("savValue"),
    form = document.getElementsByTagName('form')[0];

    form.addEventListener("submit",function(){
      sav.value = editor.getValue();
      store.set(filename, '')
    });
    var lastmodify = Date.now() - 100000;
    editor.on('change', function(e){
      var now = Date.now();
      if(now - lastmodify > 1000 * 2){
        store.set(filename, session.getValue())
        // update the saved value
        console.log('saved')
        lastmodify = now;
      }
    })
  })
})(window, document);

