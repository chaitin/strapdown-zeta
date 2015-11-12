(function (window, document) {
    console.log('!');
    var store = new Persist.Store('strapdown_editor', {swf_path: '/persist.swf'}),
        markdownEl = document.getElementsByTagName('xmp')[0] || document.getElementsByTagName('textarea')[0],
        version = markdownEl.getAttribute('version'),
        filename = window.location.pathname + "#" + version,  // # will not exist in pathname, but - is possible
        value = store.get(filename);

    //add ace
    var editor = ace.edit("editor"),
        session = editor.getSession();

    editor.setTheme("ace/theme/monokai");
    session.setMode("ace/mode/markdown");
    session.setOption("wrap", true);
    session.setTabSize(2);
    session.setUseSoftTabs(true);

    var title = document.getElementsByTagName('title')[0].innerText;
    var headlineEl = document.getElementById('headline');
    if (headlineEl) {
        headlineEl.innerHTML = title;
    }

    if (value && value != editor.getValue()) {
        if (confirm("Detected unsaved document cache, do you want to load the cache?")) {
            session.setValue(value);
        }
        store.remove(filename);
    } else {
        store.remove(filename);
    }


    //bind event
    var sav = document.getElementById("savValue"),
        form = document.getElementsByTagName('form')[0],
        toggleBtn = document.getElementsByClassName('navbar-toggle')[0],
        menus = document.getElementsByClassName('navbar-collapse')[0];

    addEvent(form, "submit", function () {
        isEditted = false;
        sav.value = editor.getValue();
        store.set(filename, session.getValue());
    });

    addEvent(toggleBtn, 'click', function () {
        var classList = menus.className.split(' ');
        classList.indexOf('collapse') > -1 ? classList.splice(classList.indexOf('collapse'), 1) : classList.push('collapse');
        menus.className = classList.join(' ');
    });

    var lastmodify = Date.now() - 2000;
    var isEditted = false;

    editor.on('change', function (e) {
        var now = Date.now();
        if (now - lastmodify > 1000 * 2) {
            store.set(filename, session.getValue());
            // update the saved value
            lastmodify = now;
        }
    });

    session.on('change', function (e) {
        isEditted = isEditted || true;
    });

    window.onbeforeunload = function () {
        store.set(filename, session.getValue());
        return isEditted ? "Document unsaved, discard changes and close window?" : void 0;
    };

    var renderedContainer = document.getElementsByClassName('render-target')[0];
    var store_of_theme = new Persist.Store('strapdown', {swf_path: '/persist.swf'});
    var theme = store_of_theme.get("theme") || "chaitin";

    var preview_toggle = document.getElementById('preview-toggle');
    addEvent(preview_toggle, 'click', function () {
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

    var uploadIframe, fileBody, fileName, uploadPath;
    var uploadBtn = document.getElementById("upload-btn");
    var uploadArea = document.getElementById("upload-area");

    function isImage(fileName) {
        var splitedName = fileName.split(".");
        if (splitedName.length > 1 && ["jpg", "jpeg", "png", "bmp", "gif"].indexOf(splitedName[splitedName.length - 1].toLowerCase()) > -1) {
            return true;
        }
        return false;
    }

    function insertContent(uploadPath) {
        if (isImage(uploadPath)) {
            // image
            editor.insert("\n\n![" + uploadPath + "](" + uploadPath + ")\n\n");
        }
        else {
            editor.insert("\n\n[" + uploadPath + "](" + uploadPath + ")\n\n");
        }
    }

    function iframeOnload() {
        if (uploadIframe.contentDocument.body.innerText == "success") {
            insertContent(uploadPath);
        }
        else {
            alert("Failed to upload");
        }
        uploadBtn.innerText = "Attach File";
        uploadIframe.removeEventListener("load", iframeOnload);
    }

    function getUploadPath(fileName) {
        var dir = location.pathname;
        return dir.slice(0, dir.lastIndexOf("/") + 1) + fileName;
    }

    uploadBtn.addEventListener("click", function () {
        uploadArea.innerHTML = '<form action="" method="post" id="upload-form" enctype="multipart/form-data">' +
            '<input type="file" id="file-body" name="body"></form>' +
            '<iframe id="upload-iframe" name="upload-iframe" style="display: none;width: 0;height: 0;tab-index: -1">';
        fileBody = document.getElementById("file-body");
        uploadIframe = document.getElementById("upload-iframe");
        fileBody.click();
        fileBody.addEventListener("change", function () {
            // auto upload
            var uploadForm = document.getElementById("upload-form");
            fileName = fileBody.value.split(/(\\|\/)/g).pop();
            uploadPath = getUploadPath(fileName);
            uploadForm.action = uploadPath + "?upload";
            uploadIframe.src = uploadPath + "?upload";
            uploadForm.target = "upload-iframe";
            uploadIframe.addEventListener("load", iframeOnload, false);
            uploadForm.submit();
            uploadBtn.innerText = "Uploading"
        }, false);
    }, false);

    var dragArea = document.getElementById("editor");

    function dragOver(e) {
        e.stopPropagation();
        e.preventDefault();
    }

    dragArea.addEventListener("dragover", dragOver, false);
    dragArea.addEventListener("dragleave", dragOver, false);
    dragArea.addEventListener("drop", function (e) {
        dragOver(e);
        var fileList = e.target.files || e.dataTransfer.files;
        var xhr = new XMLHttpRequest();
        var uploadPath = getUploadPath(fileList[0].name);
        xhr.onreadystatechange = function (e) {
            if (xhr.readyState == 4) {
                if (xhr.responseText == "success") {
                    insertContent(uploadPath);
                }
                else {
                    alert("Failed to upload");
                }
            }
        };
        xhr.open("post", uploadPath + "?upload", true);
        xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
        var formData = new FormData();
        formData.append("body", fileList[0]);
        xhr.send(formData);
    }, false);

    modals.init({backspaceClose: false});
    document.getElementById("option-btn").addEventListener("click", function () {
        modals.openModal(null, "#option-modal");
    });
    document.getElementById("option-submit-btn").addEventListener("click", function () {
        var title = document.getElementById("Title").value;
        var headingNumber = document.getElementById("HeadingNumber").value;
        var showToc = document.getElementById("Toc").checked;

        if (!title) {
            alert("Title can not be empty");
            return;
        }

        function validHeadingNumber(str) {
            // allow empty
            if(!str){
                return true;
            }
            var s = str.split(".");
            for (var i = 0; i < s.length; i++) {
                if (s[i] != "a" && s[i] != "i") {
                    return false;
                }
            }
            return true;
        }

        if (!validHeadingNumber(headingNumber)) {
            alert("Heading Number format error, it should like \"i.a'a.i\"");
            return;
        }

        var xhr = new XMLHttpRequest();
        xhr.onreadystatechange = function () {
            if (xhr.readyState == 4) {
                var response = JSON.parse(xhr.responseText);
                if (response.code) {
                    alert("Failed to save option");
                }
                else {
                    alert("Success");
                    modals.closeModals(null, "#option-modal");
                }
            }
        };
        xhr.open("POST", location.pathname + "?option");
        xhr.setRequestHeader("Content-Type", "application/json");
        if(!headingNumber){
            headingNumber = "false"
        }
        xhr.send((JSON.stringify({"Title": title, "HeadingNumber": headingNumber, "Toc": showToc.toString()})))
    })

})(window, document);
