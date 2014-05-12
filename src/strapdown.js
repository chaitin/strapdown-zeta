;(function(window, document) {

  //////////////////////////////////////////////////////////////////////
  //
  // Shims for IE < 9
  //

  document.head = document.getElementsByTagName('head')[0];

  if (!('getElementsByClassName' in document)) {
    document.getElementsByClassName = function(name) {
      function getElementsByClassName(node, classname) {
        var a = [];
        var re = new RegExp('(^| )'+classname+'( |$)');
        var els = node.getElementsByTagName("*");
        for(var i=0,j=els.length; i<j; i++)
            if(re.test(els[i].className))a.push(els[i]);
        return a;
      }
      return getElementsByClassName(document.body, name);
    }
  }


  //////////////////////////////////////////////////////////////////////
  //
  // Get user elements we need
  //

  var markdownEl = document.getElementsByTagName('xmp')[0] || document.getElementsByTagName('textarea')[0],
      titleEl = document.getElementsByTagName('title')[0],
      scriptEls = document.getElementsByTagName('script'),
      navbarEl = document.getElementsByClassName('navbar')[0];

  if (!markdownEl) {
    console.warn('No embedded Markdown found in this document for Strapdown.js to work on! Visit http://strapdownjs.com/ to learn more.');
    return;
  }

  // Hide body until we're done fiddling with the DOM
  document.body.style.display = 'none';

  //////////////////////////////////////////////////////////////////////
  //
  // <head> stuff
  //

  // Use <meta> viewport so that Bootstrap is actually responsive on mobile
  var metaEl = document.createElement('meta');
  metaEl.name = 'viewport';
  metaEl.content = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0';
  if (document.head.firstChild)
    document.head.insertBefore(metaEl, document.head.firstChild);
  else
    document.head.appendChild(metaEl);

  // Get origin of script
  var origin = '';
  for (var i = 0; i < scriptEls.length; i++) {
    if (scriptEls[i].src.match('strapdown')) {
      origin = scriptEls[i].src;
    }
  }
  var originBase = origin.substr(0, origin.lastIndexOf('/'));

  // Get theme
  var theme = markdownEl.getAttribute('theme') || 'bootstrap';
  theme = theme.toLowerCase();

  // Stylesheets
  var linkEl = document.createElement('link');
  linkEl.href = originBase + '/themes/'+theme+'.min.css';
  linkEl.rel = 'stylesheet';
  document.head.appendChild(linkEl);

  var linkEl = document.createElement('link');
  linkEl.href = originBase + '/strapdown.min.css';
  linkEl.rel = 'stylesheet';
  document.head.appendChild(linkEl);

  var linkEl = document.createElement('link');
  linkEl.href = originBase + '/themes/bootstrap-responsive.min.css';
  linkEl.rel = 'stylesheet';
  document.head.appendChild(linkEl);

  //////////////////////////////////////////////////////////////////////
  //
  // <body> stuff
  //

  var markdown = markdownEl.textContent || markdownEl.innerText;

  var newNode = document.createElement('div');
  newNode.className = 'container';
  newNode.id = 'content';
  document.body.replaceChild(newNode, markdownEl);

  // Insert navbar if there's none
  var newNode = document.createElement('div');
  newNode.className = 'navbar navbar-fixed-top';
  if (!navbarEl && titleEl) {
    newNode.innerHTML = '<div class="navbar-inner"> <div class="container"> <div id="headline" class="brand"> </div> </div> </div>';
    document.body.insertBefore(newNode, document.body.firstChild);
    var title = titleEl.innerHTML;
    var headlineEl = document.getElementById('headline');
    if (headlineEl)
      headlineEl.innerHTML = title;
  }

  //////////////////////////////////////////////////////////////////////
  //
  // Markdown!
  //

  function isMSIE() {
    var ua = window.navigator.userAgent;
    var msie = ua.indexOf('MSIE ');
    var trident = ua.indexOf('Trident/');

    if (msie > 0) {
        // IE 10 or older => return version number
      return parseInt(ua.substring(msie + 5, ua.indexOf('.', msie)), 10);
    }

    if (trident > 0) {
        // IE 11 (or newer) => return version number
        var rv = ua.indexOf('rv:');
      return parseInt(ua.substring(rv + 3, ua.indexOf('.', rv)), 10);
    }

    // other browser
    return false;
  }

  var m, i, o, l, k;
  var split_re = 3 === "aba".split(/(b)/).length ? function(a, f) {
    return a.split(f)
  } : function(a, f) {
    var b = [], c;
    if (!f.global) {
      c = f.toString();
      var d = "";
      c = c.replace(/^\/(.*)\/([im]*)$/, function(a, c, b) {
        d = b;
        return c
      });
      f = RegExp(c, d + "g")
    }
    for (var e = f.lastIndex = 0; c = f.exec(a); )
      b.push(a.substring(e, c.index)), b.push.apply(b, c.slice(1)), e = c.index + c[0].length;
    b.push(a.substring(e));
    return b
  };

  function escape_range(a, f, post_processing) {
    var c = k.slice(a, f + 1).join("").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    for (isMSIE() && (c = c.replace(/(%[^\n]*)\n/g, "$1<br/>\n")); f > a; )
      k[f] = "", f--;
    k[a] = "@@@@" + m.length + "@@@@";
    post_processing && (c = post_processing(c));
    m.push(c);
    i = o = l = null
  }

  function preprocess(raw_markdown) {
    i = o = l = null;
    m = [];
    var f;
    /`/.test(raw_markdown) ? (raw_markdown = raw_markdown.replace(/~/g, "~T").replace(/(^|[^\\])(`+)([^\n]*?[^`\n])\2(?!`)/gm, function(a) {
      return a.replace(/\$/g, "~D")
    }), f = function(a) {
      return a.replace(/~([TD])/g, 
      function(a, c) {
        return {T: "~",D: "$"}[c]
      })
    }) : f = function(a) {
      return a
    };
    k = split_re(raw_markdown.replace(/\r\n?/g, "\n"), /(\$\$?|\\(?:begin|end)\{[a-z]*\*?\}|\\[\\{}$]|[{}]|(?:\n\s*)+|@@@@\d+@@@@)/i);
    for (var a = 1, d = k.length; a < d; a += 2) {
      var c = k[a];
      if ("@" === c.charAt(0)) {
        k[a] = "@@@@" + m.length + "@@@@";
        m.push(c);
      } else {
        if (i) {
          if (c === o) {
            if (n) {
              l = a;
            } else {
              escape_range(i, a, f);
            }
          } else if (c.match(/\n/) && o == '$' ) {
            i = o = l = null;
            n = 0;
          } else {
            if (c.match(/\n.*\n/)) {
              if (l) {
                a = l;
                escape_range(i, a, f);
              }
              i = o = l = null;
              n = 0;
            } else {
              if ("{" === c) {
                n++;
              } else if ("}" === c) {
                n && n--;
              }
            }
          }
        } else {
          if (c === '$' || "$$" === c) {
            i = a; 
            o = c; 
            n = 0;
          } else if ("begin" === c.substr(1, 5)) {
            i = a; 
            o = "\\end" + c.substr(6); 
            n = 0;
          }
        }
      }
    }
    l && escape_range(i, l, f);
    ret = f(k.join(""));
    return ret;
  }

  function postprocess(a) {
    a = a.replace(/@@@@(\d+)@@@@/g, function(a, b) {
      return m[b]
    });
    m = null;
    return a
  }

  // Generate Markdown
  var markdown_without_mathjax = preprocess(markdown);
  var html = marked(markdown_without_mathjax);
  var html_with_mathjax = postprocess(html);
  document.getElementById('content').innerHTML = html_with_mathjax;

  if (html_with_mathjax != html) {
    var script = document.createElement("script");
    script.type = "text/javascript";
    script.src  = "http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS_HTML";

    var callback = function () {
      // config options
      // http://docs.mathjax.org/en/latest/options/tex2jax.html#configure-tex2jax
      MathJax.Hub.Config({
        tex2jax: {
            inlineMath: [ ['$','$']],
            displayMath:[ ['$$', '$$']],
            processEscapes: true,
            balanceBraces: true,
        },
        messageStyle: "none"
      });
      MathJax.Hub.Queue(["Typeset",MathJax.Hub]);
    }

    script.onload = callback;
    // for IE 6, IE 7
    script.onreadystatechange = function () {
      if (this.readyState == 'complete') {
        callback();
      }
    }
    document.getElementsByTagName("head")[0].appendChild(script);
  }

  // Prettify
  var codeEls = document.getElementsByTagName('code');
  for (var i=0, ii=codeEls.length; i<ii; i++) {
    var codeEl = codeEls[i];
    var lang = codeEl.className;
    codeEl.className = 'prettyprint lang-' + lang;
  }
  prettyPrint();

  // Style tables
  var tableEls = document.getElementsByTagName('table');
  for (var i=0, ii=tableEls.length; i<ii; i++) {
    var tableEl = tableEls[i];
    tableEl.className = 'table table-striped table-bordered';
  }

  // All done - show body
  document.body.style.display = '';

})(window, document);

// vim: ai:ts=2:sts=2:sw=2:
