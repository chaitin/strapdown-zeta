;(function(window, document){

  var store = new Persist.Store('strapdown_editor', { swf_path: '/persist.swf' }),
      markdownEl = document.getElementsByTagName('xmp')[0] || document.getElementsByTagName('textarea')[0],
      version = markdownEl.getAttribute('version'),
      filename = window.location.pathname + "#" + version;  // # will not exist in pathname, but - is possible

  function save(key, value){
    if(key != 'records'){
      store.get('records', function(ok, records){
        if (!ok || !records){
          records = "[]";
        }
        console.log(records)
        records = JSON.parse(records)
        if (records.slice(-1)[0] != key){
          records.push(key);
        }
        if (records.length >= 10){
          records.slice(0, records.length - 10).forEach(function(version){
            store.remove(version);
            console.log('delete the ', version);
          })
          records = records.slice(-10);
        }
        store.set(key, value)
        store.set('records', JSON.stringify(records));
      })
    }
  }

  store.get(filename, function(ok, value){
    //add ace
    var editor = ace.edit("editor"),
        session = editor.getSession();

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
      save(filename, sav.value)
    });

    var lastmodify = Date.now() - 2000;

    editor.on('change', function(e){
      var now = Date.now();
      if(now - lastmodify > 1000 * 2){
        save(filename, editor.getValue())
        // update the saved value
        lastmodify = now;
      }
    })
    document.getElementsByTagName('body')[0].addEventListener('unload',function(){
      save(filename, editor.getValue())
    })
  })
})(window, document);
