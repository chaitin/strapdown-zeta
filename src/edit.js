(function(window, document){
  console.log('!');
  var store = new Persist.Store('strapdown_editor', { swf_path: '/persist.swf' }),
      markdownEl = document.getElementsByTagName('xmp')[0] || document.getElementsByTagName('textarea')[0],
      version = markdownEl.getAttribute('version'),
      filename = window.location.pathname + "#" + version,  // # will not exist in pathname, but - is possible
      value = store.get(filename);

    //add ace
    var editor = ace.edit("editor"),
        session = editor.getSession();

    editor.setTheme("ace/theme/monokai");
    session.setMode("ace/mode/markdown");
    session.setTabSize(2);
    session.setUseSoftTabs(true);

    if (value) {
      if (confirm("Detected unsaved document cache, do you want to load the cache?")) {
        session.setValue(value);
      }
      store.clear(filename);
    }


    //bind event
    var sav = document.getElementById("savValue"),
        form = document.getElementsByTagName('form')[0],
        toggleBtn = document.getElementsByClassName('navbar-toggle')[0],
        menus = document.getElementsByClassName('navbar-collapse')[0];

    addEvent(form, "submit", function(){
      sav.value = editor.getValue();
      store.set(filename, session.getValue());
    });

    addEvent(toggleBtn, 'click', function(){
      var classList = menus.className.split(' ');
      classList.indexOf('collapse') > -1 ? classList.splice(classList.indexOf('collapse') ,1) : classList.push('collapse');
      menus.className = classList.join(' ');
    });

    var lastmodify = Date.now() - 2000;

    editor.on('change', function(e){
      var now = Date.now();
      if(now - lastmodify > 1000 * 2){
        store.set(filename, session.getValue());
        // update the saved value
        lastmodify = now;
      }
    })
    addEvent(document.getElementsByTagName('body')[0], 'unload',function(){
        store.set(filename, session.getValue());
    })

    var renderedContainer = document.getElementsByClassName('render-target')[0];
    var store_of_theme = new Persist.Store('strapdown', { swf_path: '/persist.swf' });
    var theme = store_of_theme.get("theme") || "chaitin";

    var preview_toggle = document.getElementById('preview-toggle');
    addEvent(preview_toggle, 'click', function() {
      if (renderedContainer.style.display == 'none') {
        markdownEl.style.display = 'none';
        var renderTarget = document.createElement("div"),
            markdown = session.getValue();
        renderTarget.className = 'container';
        renderTarget.id = 'content';
        renderedContainer.innerHTML = ""
        renderedContainer.appendChild(renderTarget);
        render(renderTarget, markdown, theme, null, false);
        renderedContainer.style.display = 'block';
        setInnerText(preview_toggle, "Continue Editing");
      } else {
        renderedContainer.style.display = 'none';
        markdownEl.style.display = 'block';
        setInnerText(preview_toggle, "Instant Preview");
        document.getElementsByTagName('textarea')[0].focus();
      }
    });
})(window, document);
