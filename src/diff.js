(function(window, document){
  var htmlspecialchars = function(s) {
    s = s.replace(/&/g, "&amp;");
    s = s.replace(/</g, "&lt;");
    s = s.replace(/>/g, "&gt;");
    s = s.replace(/"/g, "&quot;");
    s = s.replace(/'/g, "&#039;");
    return s;
  };

  var codeEl = document.getElementsByTagName('code')[0];
  codeEl.innerHTML = htmlspecialchars(codeEl.innerHTML);

  hljs.initHighlightingOnLoad();
})(window, document);
