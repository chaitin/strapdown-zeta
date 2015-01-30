
# Strapdown-Zeta

Strapdown-Zeta is a Git-powered wiki system, derived form strapdown.js project.

Strapdown.js makes it embarrassingly simple to create elegant Markdown documents. No server-side compilation required.


## Features

 - Git Powered Wiki system. A standalone server is provided, just `git init` then run the server will provide you a full functional geeky wiki server.
 - [MathJax](http://www.mathjax.org/) support. Feel free to type in your awesome math equations.
 - Theme switchable. 14 Bootstrap themes included by default, you can add more as you wish. And everybody can switch the theme thru one click.
 - Table of Content auto generation. Just specify `toc="true"` in the xml tag
 - Heading numbering and anchor support, just one click will bring you to the section you are going to.

For more, please see:

+ http://strapdown.ztx.io
+ [Strapdown MathJax Test Page](http://strapdown.ztx.io/test.html)
+ http://strapdownjs.com

## Usage

TODO

## Installation

### For normal users

Standalone downloadable binary will be released soon...

### For hackers

You can hack this project any way you want. Please follow the following build instructions get started.

## Build

### build strapdown.min.js and strapdown.min.css

```
$ npm install
$ grunt
```

The generated file would be `build/strapdown.min.js` and `build/strapdown.min.css`

### Build strapdown server

The server can be built into a single standalone binary.

The server is written using [go programming language](http://golang.org).

First, clone and build [git2go](https://github.com/libgit2/git2go) following the [instructions](https://github.com/libgit2/git2go#installing).

Then do the following

```
$ cd server
$ go build
```

To run this server using [systemd](https://wiki.archlinux.org/index.php/systemd), copy the [strapdown.service](server/strapdown.service) file into your /etc/systemd/system/ directory and `systemctl start strapdown`

## License

This project use [SATA License](LICENSE) (Star And Thank Author License), so you have to star this project before using. Read the [license](LICENSE) carefully.

## Credits

All credit goes to the projects below that make up most of Strapdown.js:

+ [Strapdown](http://strapdownjs.com) - Original strapdown by [r2r](http://twitter.com/r2r)
+ [MathJax](http://www.mathjax.org/) - MathJax, Beautiful math in all browsers
+ [Marked](https://github.com/chjj/marked/) - Fast Markdown parser in JavaScript
+ [Google Code Prettify](http://code.google.com/p/google-code-prettify/) - Syntax highlighting in JavaScript
+ [Highlight.js](http://highlightjs.org/) - Syntax highlighting in Javascript
+ [Twitter Bootstrap](http://twitter.github.com/bootstrap/) - Beautiful, responsive CSS framework
+ [Bootswatch](http://bootswatch.com) - Additional Bootstrap themes
+ [Stackedit](https://github.com/benweet/stackedit) - I borrowed some mathjax preprocessing code from this project. Thanks. And [https://stackedit.io/](https://stackedit.io/) is a great project!
+ [persist.js](http://pablotron.org/?cid=1557) - Client Side persistent storage solution for remembering themes.

## Contributor

 - [zTrix](https://github.com/zTrix)
 - [cbmixx](https://github.com/cbmixx)
 - [qoshi](https://github.com/qoshi)
 - [arturadib](https://github.com/arturadib)

