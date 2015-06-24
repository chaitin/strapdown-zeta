;(function(window, document) {

var store = new Persist.Store('strapdown', { swf_path: '/persist.swf' });

store.get('theme', function (ok, val) {
  var theme = null;
  if (ok) {
    theme = val;
  }

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
    console.warn('No embedded Markdown found in this document for Strapdown.js to work on! Visit http://strapdown.ztx.io/ to learn more.');
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
  theme = theme || markdownEl.getAttribute('theme') || 'cerulean';
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
    newNode.innerHTML = '<div class="navbar-inner"> <div class="container">' + 
                        '<a class="btn btn-navbar" data-toggle="collapse" data-target=".navbar-responsive-collapse"><span class="icon-bar"></span><span class="icon-bar"></span><span class="icon-bar"></span></a>' +
                        '<div id="headline" class="brand"> </div>' +
                        '<div class="nav-collapse collapse navbar-responsive-collapse pull-right"> <a class="btn btn-default navbar-btn" href="?history">History</a> <a class="btn btn-default navbar-btn" href="?edit">Edit</a> <ul class="nav"><li class="dropdown"><a href="javascript:;" class="dropdown-toggle" data-toggle="dropdown">Theme<b class="caret"></b></a><ul class="dropdown-menu" id="theme"></ul></li></ul> </div>' +

                        '</div> </div>';
    document.body.insertBefore(newNode, document.body.firstChild);
    var title = titleEl.innerHTML;
    var headlineEl = document.getElementById('headline');
    if (headlineEl) {
      headlineEl.innerHTML = title;
    }
    function addEvent(element, evnt, funct) {
      if (element.attachEvent)
       return element.attachEvent('on'+evnt, funct);
      else
       return element.addEventListener(evnt, funct, false);
    }

    var themeEl = document.getElementById('theme');
    if (themeEl) {
      var themes = ['Amelia', 'Bootstrap', 'Cerulean', 'Cosmo', 'Cyborg', 'Flatly', 'Journal', 'Readable', 'Simplex', 'Slate', 'Spacelab', 'Spruce', 'Superhero', 'United', 'Reset'];
      themes.forEach(function(val) {
        if (val == 'Reset') {
          var dvd = document.createElement("li");
          dvd.setAttribute("class", "divider");
          themeEl.appendChild(dvd);
        }
        var li = document.createElement("li");
        var a = document.createElement("a");
        if (typeof a.textContent !== 'undefined') {
          a.textContent = val;
        } else {
          a.innerText = val;
        }
        a.setAttribute('href', '#');
        li.appendChild(a);
        addEvent(a, 'click', function () {
          if (val == 'Reset') {
            store.set('theme', '');
          } else {
            store.set('theme', val);
          }
          location.reload();
        });
        themeEl.appendChild(li);
      });
    }
    var dropdown = document.getElementsByClassName("dropdown")[0];
    if (themeEl && dropdown) {
      addEvent(dropdown, 'click', function () {
        // console.log('click dropdown', dropdown.className.match(/(?:^|\s)open(?!\S)/));
        if (dropdown.className.match(/(?:^|\s)open(?!\S)/)) {
          dropdown.className = dropdown.className.replace(/(?:^|\s)open(?!\S)/g, '');
        } else {
          dropdown.className += " open";
        }
      });
    }
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

  // From math.stackexchange.com...
  // borrowed from https://github.com/benweet/stackedit, thanks
  // https://stackedit-beta.herokuapp.com/res/extensions/mathJax.js

  //
  //  The math is in blocks i through j, so
  //    collect it into one block and clear the others.
  //  Replace &, <, and > by named entities.
  //  For IE, put <br> at the ends of comments since IE removes \n.
  //  Clear the current math positions and store the index of the
  //    math, then push the math string onto the storage array.
  //
  var blocks, start, end, last, braces, math;

  function processMath(i, j, unescape) {
    var block = blocks.slice(i, j + 1).join("")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
    for(isMSIE() && (block = block.replace(/(%[^\n]*)\n/g, "$1<br/>\n")); j > i;)
      blocks[j] = "", j--;
    blocks[i] = "@@@@" + math.length + "@@@@";
    unescape && (block = unescape(block));
    math.push(block);
    start = end = last = null;
  }

  function removeMath(text) {
    start = end = last = null;
    math = [];
    var unescape;
    if(/`/.test(text)) {
      text = text.replace(/~/g, "~T").replace(/(^|[^\\])(`+)([^\n]*?[^`\n])\2(?!`)/gm, function(text) {
        return text.replace(/\$/g, "~D")
      });
      unescape = function(text) {
        return text.replace(/~([TD])/g,
          function(match, n) {
            return {T: "~", D: "$"}[n]
          })
      };
    } else {
      unescape = function(text) {
        return text
      };
    }
    blocks = split(text.replace(/\r\n?/g, "\n"), splitDelimiter);
    for(var i = 1, m = blocks.length; i < m; i += 2) {
      var block = blocks[i];
      if("@" === block.charAt(0)) {
        //
        //  Things that look like our math markers will get
        //  stored and then retrieved along with the math.
        //
        blocks[i] = "@@@@" + math.length + "@@@@";
        math.push(block)
      } else if(start) {
        // Ignore inline maths that are actually multiline (fixes #136)
        if (end == '$' && block.match(/\n/)) {
          if(last) {
            i = last;
            processMath(start, i, unescape);
          }
          start = end = last = null;
          braces = 0;
        }
        //
        //  If we are in math, look for the end delimiter,
        //    but don't go past double line breaks, and
        //    and balance braces within the math.
        //
        else if(block === end) {
          if(braces) {
            last = i
          } else {
            processMath(start, i, unescape)
          }
        } else {
          if(block.match(/\n.*\n/)) {
            if(last) {
              i = last;
              processMath(start, i, unescape);
            }
            start = end = last = null;
            braces = 0;
          } else {
            if("{" === block) {
              braces++
            } else {
              "}" === block && braces && braces--
            }
          }
        }
      } else {
        if(block === '$' || "$$" === block) {
          start = i;
          end = block;
          braces = 0;
        } else {
          if("begin" === block.substr(1, 5)) {
            start = i;
            end = "\\end" + block.substr(6);
            braces = 0;
          }
        }
      }

    }
    last && processMath(start, last, unescape);
    return unescape(blocks.join(""))
  }

  //
  //  Put back the math strings that were saved,
  //    and clear the math array (no need to keep it around).
  //
  function replaceMath(text) {
    text = text.replace(/@@@@(\d+)@@@@/g, function(match, n) {
      return math[n]
    });
    math = null;
    return text
  }

  //
  //  The pattern for math delimiters and special symbols
  //    needed for searching for math in the page.
  //
  var splitDelimiter = /(\$\$?|\\(?:begin|end)\{[a-z]*\*?\}|\\[\\{}$]|[{}]|(?:\n\s*)+|@@@@\d+@@@@)/i;
  var split;

  if(3 === "aba".split(/(b)/).length) {
    split = function(text, delimiter) {
      return text.split(delimiter)
    };
  } else {
    split = function(text, delimiter) {
      var b = [], c;
      if(!delimiter.global) {
        c = delimiter.toString();
        var d = "";
        c = c.replace(/^\/(.*)\/([im]*)$/, function(a, c, b) {
          d = b;
          return c
        });
        delimiter = RegExp(c, d + "g")
      }
      for(var e = delimiter.lastIndex = 0; c = delimiter.exec(text);) {
        b.push(text.substring(e, c.index));
        b.push.apply(b, c.slice(1));
        e = c.index + c[0].length;
      }
      b.push(text.substring(e));
      return b
    };
  }

  var toc = [];
  var heading_counter = [0, 0, 0, 0, 0, 0];

  var heading_number = markdownEl.getAttribute('heading_number');

  var hn_table = ['i', 'i', 'i', 'i', 'i', 'i'];
  if (heading_number && heading_number != 'none' && heading_number != "false" ) {
    var ary = heading_number.split('.');
    for (var i = 0; i < 6; i++) {
      if (ary[i] == 'a') {
        hn_table[i] = 'a';
      }
    }
  }

  var itoa = function (i, j) {
    if (hn_table[j] == 'a' && i <= 26) {
      return String.fromCharCode(96 + i);
    } else {
      return '' + i;
    }
  }

  var counter_to_str = function (hc) {
    var i = 5;
    var ret = "" + itoa(hc[0], 0);
    for (; i >= 0; i--) {
      if (hc[i]) break;
    }
    for (var j = 1; j <= i; j++) {
      ret += "." + itoa(hc[j], j);
    }
    return ret;
  };

  var toc = [];

  var renderer = new marked.Renderer();
  renderer.heading = function (text, level) {

    heading_counter[level-1]++;
    for (var i = level; i < 6; i++) {
      heading_counter[i] = 0;
    }

    var heading_number_str = counter_to_str(heading_counter);

    var escapedText = 'h' + heading_number_str + '_' + text.toLowerCase().replace(/[^-_.\w\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF]+/g, '-');

    // generate heading
    var before_heading;
    if (!heading_number || heading_number == 'none' || heading_number == "false") {
      before_heading = '';
    } else {
      before_heading = heading_number_str + ' ';
    }

    // for table of content
    var a = toc;
    for (var i = 0; i < level-1; i++) {
      if (a.length == 0 || !Array.isArray(a[a.length-1])) {
        a.push([]);
      }
      a = a[a.length-1];
    }
    a.push({
      'target': '#' + escapedText,
      'title': before_heading + text
    });

    return '<h' + level + ' style="position:relative;"><a name="' +
                escapedText +
                 '" class="anchor" href="#' +
                 escapedText +
                 '"><span class="header-link"></span></a>' + before_heading + 
                  text + '</h' + level + '>';
  }

  // Generate Markdown
  var markdown_without_mathjax = removeMath(markdown);
  var html = marked(markdown_without_mathjax, { renderer: renderer } );

  var html_with_mathjax = replaceMath(html);

  var show_toc = markdownEl.getAttribute('toc');
  if (show_toc == 'true') {
    var toc_html = document.createElement('ul');
    
    var traverse = function(list, ul) {
      for (var i = 0; i < list.length; i++) {
        var e;
        if (Array.isArray(list[i])) {
          e = document.createElement('ul');
          traverse(list[i], e);
        } else {
          e = document.createElement('li');
          var a = document.createElement('a');
          a.setAttribute('href', list[i].target);
          a.appendChild(document.createTextNode(list[i].title));
          e.appendChild(a);
        }
        ul.appendChild(e);
      }
    }
    traverse(toc, toc_html);

    var div = document.createElement('div');
    div.className = 'container';
    var title = document.createElement('h1');
    title.appendChild(document.createTextNode('Table of Content'));
    div.appendChild(title);
    div.appendChild(document.createElement('hr'));
    div.appendChild(toc_html);
    div.appendChild(document.createElement('hr'));
    document.body.insertBefore(div, document.getElementById('content'));
  }

  document.getElementById('content').innerHTML = html_with_mathjax;

  if (html_with_mathjax != html) {
    var script = document.createElement("script");
    script.type = "text/javascript";
    script.src  = "http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_SVG";

    var callback = function () {
      // config options
      // http://docs.mathjax.org/en/latest/options/tex2jax.html#configure-tex2jax
      MathJax.Ajax.timeout = 60000;
      MathJax.Hub.Config({
        tex2jax: {
            inlineMath: [ ['$','$']],
            displayMath:[ ['$$', '$$']],
            processEscapes: true,
            balanceBraces: true,
        },
        messageStyle: "none",
        SVG: {
          styles: {
            ".MathJax_SVG svg > g, .MathJax_SVG_Display svg > g": {
              "fill": "#4d4d4c",
              "stroke": "#4d4d4c"
            }
          },
          scale: 100
        }
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

  if ('hljs' in window) {
    var codeEls = document.getElementsByTagName('code');
    for (var i=0, ii=codeEls.length; i<ii; i++) {
      var codeEl = codeEls[i];
      var lang = codeEl.className;
      if (codeEl.parentNode.nodeName.toLowerCase() == 'pre') {
        codeEl.parentNode.className = 'code-wrapper ' + lang;
      }
    }
    hljs.initHighlightingOnLoad();
  } else if ('prettyPrint' in window) {
    // Prettify
    var codeEls = document.getElementsByTagName('code');
    for (var i=0, ii=codeEls.length; i<ii; i++) {
      var codeEl = codeEls[i];
      var lang = codeEl.className;
      if (codeEl.parentNode.nodeName.toLowerCase() == 'pre') {
        codeEl.parentNode.className = 'code-wrapper prettyprint ' + lang;
      }
    }
    prettyPrint();
  }

  // Style tables
  var tableEls = document.getElementsByTagName('table');
  for (var i=0, ii=tableEls.length; i<ii; i++) {
    var tableEl = tableEls[i];
    tableEl.className = 'table table-striped table-bordered';
  }

  // All done - show body
  document.body.style.display = '';

});

})(window, document);

// vim: ai:ts=2:sts=2:sw=2:
