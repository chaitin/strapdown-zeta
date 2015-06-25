;(function(window, document){

  var store = new Persist.Store('strapdown_editor', { swf_path: '/persist.swf' }),
      markdownEl = document.getElementsByTagName('xmp')[0] || document.getElementsByTagName('textarea')[0],
      version = markdownEl.getAttribute('version'),
      filename = window.location.pathname + "#" + version;  // # will not exist in pathname, but - is possible

  store.get(filename, function(ok, value){

    //add ace
    var editor = ace.edit("editor"),
        session = editor.getSession(),
        saved = false;

    editor.setTheme("ace/theme/monokai");
    session.setMode("ace/mode/markdown");
    session.setTabSize(2);
    session.setUseSoftTabs(true);

    if (ok && value) {
      if (confirm("Detected unsaved document cache, do you want to load the cache?")) {
        session.setValue(value)
      }
    }


    //bind event
    var sav = document.getElementById("savValue"),
        form = document.getElementsByTagName('form')[0];

    form.addEventListener("submit",function(){
      sav.value = editor.getValue();
      store.set(filename, session.getValue())
      saved = true
    });

    var lastmodify = Date.now() - 2000;

    editor.on('change', function(e){
      var now = Date.now();
      if(now - lastmodify > 1000 * 2){
        store.set(filename, session.getValue())
        // update the saved value
        lastmodify = now;
      }
    })
    document.getElementsByTagName('body')[0].addEventListener('unload',function(){
      if (!saved) {
        store.set(filename, session.getValue());
      }
    })

    var renderedContainer = document.getElementsByClassName('render-target')

    document.getElementById('preview-toggle').addEventListener('click', function(){
      if(renderedContainer.styles.display == 'none'){
        renderedContainer.innerHtml = "<xmp>" + session.getValue() + "<xmp>";
        // and call the render stuff here
        renderedContainer.styles.display = 'block';
        markdownEl.styles.display = 'none';
      }else{
        renderedContainer.styles.display = 'none';
        markdownEl.styles.display = 'block';
      }
    })
  })
})(window, document);
