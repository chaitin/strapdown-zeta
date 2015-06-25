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

    var renderedContainer = document.getElementsByClassName('render-target')[0]

    document.getElementById('preview-toggle').addEventListener('click', function(){
      console.log("You Clicl the preview page")
      if(renderedContainer.style.display == 'none'){
        var renderTarget = document.createElement("div"),
            markdown = session.getValue();
        render(renderTarget, markdown, 'cerulean', null, false)
        renderedContainer.innerHTML = ""
        renderedContainer.appendChild(renderTarget);
        renderedContainer.style.display = 'block';
        markdownEl.style.display = 'none';
      }else{
        renderedContainer.style.display = 'none';
        markdownEl.style.display = 'block';
      }
    })
  })
})(window, document);
