
# Strapdown-Zeta - Git powered Wiki for Hackers!

Strapdown-Zeta is a git-powered wiki system for hackers, derived from [strapdown.js](http://strapdownjs.com/) project.

Strapdown.js makes it embarrassingly simple to create elegant Markdown documents. No server-side compilation required.

Strapdown-Zeta add more features including a standalone server providing a git powered wiki system, on top of [libgit2](https://github.com/libgit2/git2go), we don't want any potential command injections!

## Features

### Strapdown Static Markdown Page

 - Search Engine friendly.
 - Cross-browser compatible and responsive in mobile screens.
 - GitHub flavored Markdown syntax.
 - [MathJax](http://www.mathjax.org/) support. Feel free to type in your awesome math equations.
 - Theme switchable. 14 Bootstrap themes included by default, you can add more as you wish. And everybody can switch the theme thru one click.
 - `Table of Content` auto generation. Just specify `toc="true"` in the `xmp` tag
 - Heading numbering and anchor support, just one click will bring you to the section you are going to.
 - Use highlight.js for syntax highlighting, which provides more beautiful coloring and more powerful syntax parsing.
 - Blazing fast loading speed! All the codes are written using [vanilla-js](http://vanilla-js.com/), only less than 200KiB source code after compressing.

### Git powered Wiki

 - Git Powered Wiki system. A standalone server is provided, just `git init` then run the server will get you a full functional geeky wiki server.
 - File modification history and view by commit version(shortened sha hash).
 - Custom view options can be specified for different files.
 - Handle of static files. Directory listing can be turned on and off.
 - HTTP Authentication provided.

For more, please see:

+ http://strapdown.ztx.io
+ [Strapdown MathJax Test Page](http://strapdown.ztx.io/test.html)

## Usage

### Use Strapdown static html

Open your favorite text editor, paste the following lines into the text, then type markdown content in the middle, save the file as test.html, and open it, here you go!

```
<!DOCTYPE html> <html> <title>Hello, Strapdown</title> <meta charset="utf-8"> <xmp theme="cerulean" style="display:none;">

# title

your awesome markdown content goes here...

</xmp> <script src="http://cdn.ztx.io/strapdown/strapdown.min.js"></script> </html>
```

#### Choose theme

You can set your favorite theme in `xmp` tag. The following themes are built-in by default.

 - Amelia
 - Bootstrap
 - Cerulean
 - Cosmo
 - Cyborg
 - Flatly
 - Journal
 - Readable
 - Simplex
 - Slate
 - Spacelab
 - Spruce
 - Superhero
 - United

To use Cosmo, use the following line

```
<!DOCTYPE html> <html> <title>Hello, Strapdown</title> <meta charset="utf-8"> <xmp theme="cosmo" style="display:none;">

your awesome markdown content goes here...

</xmp> <script src="http://cdn.ztx.io/strapdown/strapdown.min.js"></script> </html>
```

#### Table of Content

To generate table of content, specify `toc="true"` in xmp tag.

```
<!DOCTYPE html> <html> <title>Hello, Strapdown</title> <meta charset="utf-8"> <xmp theme="cosmo" toc="true" style="display:none;">

your awesome markdown content goes here...

</xmp> <script src="http://cdn.ztx.io/strapdown/strapdown.min.js"></script> </html>
```

### Strapdown Server

The server supports the following parameters.

 - `-addr="0.0.0.0"`, specify the listening host:port tuple, multiple addresses can be specified by separation of comma, e.g. `192.168.1.10:8080,127.0.0.1:8080`.
 - `-init`, do automatic `git init` before starting the server, if git repo not found in working directory.
 - `-dir=/path/to/dir`, use the directory as the root of the git powered wiki.
 - `-title=MyTitle`, specify the default title of Wiki
 - `-auth=.htpasswd`, specify the authentication file to use, htpasswd format
 - `-heading_number=true|false`, set default value for whether to show heading numbers
 - `-host=some.domain.com`, the default hosting of strapdown static files
 - `-theme=cerulean|cosmo|...`, the default theme to use

## Installation

### For normal users

Standalone downloadable binary will be released soon...

### For hackers

You can hack this project any way you want. Please follow the following build instructions to get started.

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
$ go get
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
 - [zyaboutblank](https://github.com/zyaboutblank)
 - [pandada8](https://github.com/pandada8)

