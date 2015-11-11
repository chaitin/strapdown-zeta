package main

import (
	"errors"
	"flag"
	"fmt"
	auth "github.com/abbot/go-http-auth"
	"github.com/libgit2/git2go"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"encoding/json"
)

type DirEntry struct {
	Urlpath string
	Name    string
	Size    int64
	IsDir   bool
	ModTime time.Time
}

type CommitEntry struct {
	Id        string
	EntryId   string
	Timestamp time.Time
	Author    string
	Message   string
}

func (this *CommitEntry) ShortHash() string {
	return this.Id[:11]
}

type Config struct {
	addr           string
	init           bool
	root           string
	auth           string
	host           string
	heading_number string
	title          string
	theme          string
	histsize       int
	toc            string
	verbose        bool
	version        bool
	optext         string
	extract        bool
	prefix         string
}

type RequestContext struct {
	Title         string
	Theme         string
	Toc           string
	HeadingNumber string
	Content       template.HTML
	DirEntries    []DirEntry
	CommitEntries []CommitEntry
	Version       string
	Host          string //deleteme

	path       string
	res        *http.ResponseWriter
	req        *http.Request
	ip         string
	isFolder   bool
	username   string
	statusCode int
}

type CustomOption struct {
	Title         string
	Theme         string
	Toc           string
	HeadingNumber string
	Host          string
}

var wikiConfig Config // the global config file
var templates map[string]*template.Template
var authenticator *auth.BasicAuth

var SERVER_VERSION string

func parseConfig() {
	flag.StringVar(&wikiConfig.addr, "addr", ":8080", "Listening `host:port`, you can specify multiple listening address separated by comma, e.g. (127.0.0.1:8080,192.168.1.2:8080)")
	flag.BoolVar(&wikiConfig.init, "init", false, "init git repository before running, just like `git init`")
	flag.StringVar(&wikiConfig.root, "dir", "", "The root directory for the git/wiki")
	flag.StringVar(&wikiConfig.auth, "auth", ".htpasswd", "Default auth file to use as authentication, authentication will be disabled if auth file not exist")
	flag.StringVar(&wikiConfig.host, "host", "/_static", "URL prefix where host hosting the strapdown static files")
	flag.StringVar(&wikiConfig.heading_number, "heading_number", "false", "set default value for showing heading number")
	flag.StringVar(&wikiConfig.title, "title", "Wiki", "default title for wiki pages")
	flag.StringVar(&wikiConfig.theme, "theme", "chaitin", "default theme for strapdown")
	flag.IntVar(&wikiConfig.histsize, "histsize", 30, "default history size")
	flag.StringVar(&wikiConfig.toc, "toc", "false", "set default value for showing table of content")
	flag.BoolVar(&wikiConfig.verbose, "verbose", false, "be verbose")
	flag.BoolVar(&wikiConfig.version, "v", false, "show version")
	flag.BoolVar(&wikiConfig.version, "version", false, "show version")
	flag.StringVar(&wikiConfig.optext, "optext", ".option.json", "set the option filename extension")
	flag.BoolVar(&wikiConfig.extract, "extract", false, "Extract assets to current working directory")
	flag.StringVar(&wikiConfig.prefix, "prefix", "", "Use your own static files. Unless you know what you are doing, don't use this option with -host.")
	flag.Parse()
}

func (this *DirEntry) ReadableSize(use_kibibyte bool) string {
	num := float32(this.Size)
	base := float32(1000)
	unit := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	if use_kibibyte {
		base = float32(1024)
		unit = []string{"B", "kiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	}
	var cur string
	for _, x := range unit {
		if -base < num && num < base {
			return fmt.Sprintf("%3.1f %s", num, x)
		}
		num = num / base
		cur = x
	}
	return fmt.Sprintf("%3.1f %s", num, cur)
}

// copied from http://golang.org/src/net/http/fs.go
func SafeOpen(base string, name string) (*os.File, error) {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := base
	if dir == "" {
		dir = "."
	}
	f, err := os.Open(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getHeadVersion() string {
	repo, err := git.OpenRepository(".")
	if err != nil {
		return ""
	}
	head, err := repo.Head()
	repo.Free() // no matter err is or isnot nil, free repo
	if err != nil {
		return ""
	}
	return head.Target().String()
}

func bootstrap() {

	mime.AddExtensionType(".md", "text/markdown")

	v, err := Asset("_static/version")
	if err != nil {
		log.Printf("[ WARN ] server version not found, wrong build")
	} else {
		SERVER_VERSION = strings.TrimSpace(string(v))
	}

	if len(wikiConfig.root) > 0 {
		// we should chdir to the root
		err := os.Chdir(wikiConfig.root)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		log.Printf("chdir to the '%s'", wikiConfig.root)
	}

	if wikiConfig.init {
		if repo, err := git.OpenRepository("."); err != nil {
			_, err := git.InitRepository(".", false)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			log.Printf("git init finished at .")
		} else {
			log.Printf("git repository already found, skip git init")
			repo.Free()
		}
	}

	if wikiConfig.version {
		fmt.Printf("Strapdown Wiki Server - v%s\n", SERVER_VERSION)
		os.Exit(0)
	}

	pages := []string{"view", "listdir", "history", "diff", "edit", "upload"}
	templates = make(map[string]*template.Template)

	if len(wikiConfig.prefix) > 0 {
		var data []byte
		for _, element := range pages {
			data, err = ioutil.ReadFile(filepath.Join(wikiConfig.prefix, element+".html"))
			if err != nil {
				log.Fatalf("fail to load the %s.html", element)
			}
			templates[element], err = template.New(element).Parse(string(data))
			if err != nil {
				log.Fatalf("cannot parse %s template, %s", element, err)
			}
		}
	} else {
		// load template from assets
		for _, element := range pages {
			data, err := Asset("_static/" + element + ".html")
			if err != nil {
				log.Fatalf("fail to load the %s.html", element)
			}
			templates[element], err = template.New(element).Parse(string(data))
			if err != nil {
				log.Fatalf("cannot parse %s template, %s", element, err)
			}
		}
	}

	if wikiConfig.extract {
		// extract and exit
		files := AssetNames()

		for _, name := range files {
			if strings.EqualFold(name, "_static/.md") {
				// not to release the default markdown
				continue
			}
			file, err := Asset(name)
			if err != nil {
				log.Printf("[ WARN ] fail to load: %s", name)
			}
			err = os.MkdirAll(path.Dir(name), 0700)
			if err != nil {
				log.Printf("[ WARN ] fail to create folder: %s", path.Dir(name))
			}
			err = ioutil.WriteFile(name, file, 0644)
			if wikiConfig.verbose {
				log.Printf("[ DEBUG ] create: %s", name)
			}
			if err != nil {
				log.Printf("[ WARN ] cannot write file: %v", err)
			}
		}
		log.Printf("[ INFO ] Assets Extracted, exited.")
		os.Exit(0)
	}
}

func (this *RequestContext) parseIp() {
	this.ip = this.req.RemoteAddr
	i := strings.IndexByte(this.ip, ':')
	if i > -1 {
		this.ip = this.ip[:i]
	}
	if this.req.Header.Get("X-FORWARDED-FOR") != "" {
		if strings.Index(this.ip, "127.0.0.1") == 0 {
			this.ip = this.req.Header.Get("X-FORWARDED-FOR")
		} else {
			this.ip = fmt.Sprintf("%s,%s", this.ip, this.req.Header.Get("X-FORWARDED-FOR"))
		}
	}
}

// this handleFunc parse request and parameters, then dispatch the action to action.go
func handleFunc(w http.ResponseWriter, r *http.Request) {
	var err error

	var ctx RequestContext
	ctx.req = r
	ctx.res = &w
	// init to 200 OK, if no error happens, then 200 will be printed by log
	ctx.statusCode = http.StatusOK
	ctx.Title = wikiConfig.title
	ctx.Theme = wikiConfig.theme
	ctx.Toc = wikiConfig.toc
	ctx.HeadingNumber = wikiConfig.heading_number
	ctx.Host = wikiConfig.host

	w.Header().Set("X-Powered-By", "Strapdown Server (v"+SERVER_VERSION+")")

	defer func() {
		if !wikiConfig.verbose {
			log.Printf("[ %s ] - %d %s", r.Method, ctx.statusCode, r.URL.String())
		} else {
			log.Printf("[ %s ] - %d %s (%s,%s)", r.Method, ctx.statusCode, r.URL.String(), ctx.path, w.Header().Get("Content-Type"))
		}
	}()

	// check auth first
	if authenticator != nil { // check http auth
		if ctx.username = authenticator.CheckAuth(r); ctx.username == "" {
			ctx.statusCode = http.StatusUnauthorized // we need to setup statuscode every return to enable defered log to work
			authenticator.RequireAuth(w, r)
			return
		}
	}

	// parse info from parameter first
	ctx.parseIp()

	var param_version string = ""

	fp := r.URL.Path[1:]
	fpmd := fp + ".md"
	fpstat, fperr := os.Stat(fp)
	fpmdstat, fpmderr := os.Stat(fpmd)

	// consider the situation that, the xx.md file does not exist, but does exist in some history version
	// the following `if` will fail, but luckily, we can still use xx.md?history to view the history
	// so the following logic works
	if fpmderr == nil && !fpmdstat.IsDir() {
		ctx.path = fpmd
	} else {
		ctx.path = fp
	}

	// forbidden any access of git related object
	if strings.HasPrefix(fp, ".git/") || fp == ".git" || fp == ".gitignore" || fp == ".gitmodules" {
		ctx.statusCode = http.StatusForbidden
		http.Error(w, "access of .git related files/directory not allowed", ctx.statusCode)
		return
	}
	// forbidden any access of auth related object
	if len(wikiConfig.auth) > 0 && fp == wikiConfig.auth {
		ctx.statusCode = http.StatusForbidden
		http.Error(w, "access of password file not allowed", ctx.statusCode)
		return
	}

	// cache is evil
	if r.Method == "GET" {
		if strings.HasPrefix(fp, "_static") || strings.HasSuffix(fp, "favicon.ico") {
			w.Header().Set("Cache-Control", "max-age=86400, public")
		} else {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, post-check=0, pre-check=0, max-age=0")
			w.Header().Set("Expires", "Sun, 19 Nov 1978 05:00:00 GMT")
		}
	}
	if strings.HasPrefix(fp, "_static") {
		// when deal with _static, only get is allowed.
		if r.Method != "GET" {
			ctx.statusCode = http.StatusMethodNotAllowed
			http.Error(w, "Only GET to _static is allowed", ctx.statusCode)
			return
		}

		var mimetype string
		var asset []byte
		if len(wikiConfig.prefix) == 0 {
			// assets prefered
			asset, err = Asset(fp)
			if err != nil && fperr == nil && !fpstat.IsDir() {
				asset, err = ioutil.ReadFile(fp)
			}
		} else {
			var f *os.File
			f, err = SafeOpen(wikiConfig.prefix, fp[7:]) // without the _static
			f.Close()
			if err == nil {
				asset, err = ioutil.ReadFile(f.Name())
			}
		}
		if err != nil {
			ctx.statusCode = http.StatusInternalServerError
			http.Error(w, "http: "+err.Error(), ctx.statusCode)
			return
		}

		lastdot := strings.LastIndex(fp, ".")
		if lastdot > -1 {
			mimetype = mime.TypeByExtension(fp[lastdot:])
		}

		if mimetype == "" {
			if len(asset) == 0 {
				mimetype = "text/plain"
			} else {
				mimetype = http.DetectContentType(asset)
			}
		}
		w.Header().Set("Content-Type", mimetype)

		w.Write(asset)
		return
	}

	q := r.URL.Query()

	histsize_ary, dohistory := q["history"]
	diff_ary, dodiff := q["diff"]
	_, dooption := q["option"]
	_, dodelete := q["delete"]

	edit_ary, doedit := q["edit"]
	version_ary, doversion := q["version"]

	_, doupload := q["upload"]

	// version is not a standalone action
	// it can be bound to edit or view actions, but history, diff, option just ignore version param
	// so we parse versions first
	ctx.Version = getHeadVersion()
	if doversion {
		if len(version_ary) > 0 && len(version_ary[0]) > 0 {
			// note that
			// this.Version is for View/Edit template
			// param_version is the param user requested
			param_version = version_ary[0]
			ctx.Version = param_version
		} else {
			// default to latest
			// that is, if the URL is http://wiki/xxx?version it is the same as http://wiki/xxx
			param_version = ""
		}
	}

	if dohistory {
		if r.Method != "GET" {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, r.Method+" method not allowed for history", ctx.statusCode)
			return
		}
		histsize := wikiConfig.histsize
		if len(histsize_ary) > 0 {
			histsize, err = strconv.Atoi(histsize_ary[0])
			if err != nil {
				histsize = wikiConfig.histsize
			}
		}

		err = ctx.History(histsize)
		if err != nil {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, err.Error(), ctx.statusCode)
		}
		return
	}

	if dodiff {
		if r.Method != "GET" {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, r.Method+" method not allowed for diff", ctx.statusCode)
			return
		}
		if len(diff_ary) == 0 {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, "params required for diff", ctx.statusCode)
			return
		}
		diff_param := diff_ary[0]
		diff_parts := strings.Split(diff_param, ",")

		err = ctx.Diff(diff_parts)
		if err != nil {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, err.Error(), ctx.statusCode)
		}
		return
	}

	if dooption {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")

			decoder := json.NewDecoder(ctx.req.Body)
			var option CustomOption
			err := decoder.Decode(&option)

			if err != nil || option.Title == "" || (option.Toc != "true" && option.Toc != "false") {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("{\"code\": 1}"))
				return
			}

			if option.HeadingNumber != "false" {
				s := strings.Split(option.HeadingNumber, ".")
				for i:=0; i<len(s); i++ {
					if s[i] != "a" && s[i] != "i" {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("{\"code\": 1}"))
						return
					}
				}
			}
			if ctx.saveOption(option) != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("{\"code\": 1}"))
			} else {
				w.Write([]byte("{\"code\": 0}"))
			}
		} else {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, r.Method+" method not allowed for option", ctx.statusCode)
		}
		return
	}

	if dodelete {
		if r.Method == "GET" {
			// TODO: return delete template, confirm for delete operation
		} else if r.Method == "POST" || r.Method == "DELETE" {
			// TODO: delete operation
			// if the file is in git, delete and commit
			// else, just delete that file
			// but, _static/ folder and favicon.ico should not be deleted
		} else {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, r.Method+" method not allowed for option", ctx.statusCode)
			return
		}
		return
	}

	if doedit {
		// this edit function is just for edit of .md files
		// so set path to fpmd
		ctx.path = fpmd
		if r.Method == "GET" {
			if fperr == nil && !fpstat.IsDir() {
				if len(edit_ary) > 0 && edit_ary[0] == "raw" {
					// if the user set edit=raw
					// then use the original fp
					ctx.path = fp
				} else {
					ctx.statusCode = http.StatusBadRequest
					http.Error(w, fmt.Sprintf("file %s exists, use ?edit=raw", fp), ctx.statusCode)
					return
				}
			}
			// will return edit template
			err = ctx.Edit(param_version)
		} else if r.Method == "POST" || r.Method == "PUT" {
			err = ctx.Update("redirect")
		} else {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, r.Method+" method not allowed for edit", ctx.statusCode)
			return
		}
		if err != nil {
			ctx.statusCode = http.StatusBadRequest
			http.Error(w, err.Error(), ctx.statusCode)
		}
		return
	}

	if doupload {
		if r.Method == "GET" {
			err = ctx.Upload()
			if err != nil {
				ctx.statusCode = http.StatusBadRequest
				http.Error(w, err.Error(), ctx.statusCode)
			}
			return
		}
	}

	// finally, when no option provided
	// View/Update/ListDir according to http method and fs state

	err = nil

	if r.Method == "POST" || r.Method == "PUT" {
		// no edit, so upload to fp
		ctx.path = fp
		if doupload {
			err = ctx.Update("show_result")
		} else {
			err = ctx.Update("redirect")
		}
	} else if r.Method == "GET" {
		if fperr == nil { // fp exists
			if fpstat.IsDir() { // fp is a dir
				if !strings.HasSuffix(fp, "/") { // redirect
					err = ctx.Redirect(r.URL.Path + "/")
				} else if fpmderr == nil && !fpmdstat.IsDir() { // .md exists, dont list dir
					ctx.path = fpmd
					err = ctx.View(param_version)
				} else if len(fp) == 0 { // root directory, dont list, just goto edit (cuz fpmd does not exists)
					ctx.path = fpmd
					err = ctx.Edit(param_version)
				} else { // now we can listdir
					err = ctx.Listdir()
				}
			} else { // host static file
				ctx.path = fp
				ctx.Static(param_version)
			}
		} else if fpmderr == nil { // fpmd exists, just view
			if fpmdstat.IsDir() { // sadly, fpmd is a directory, show error
				ctx.statusCode = http.StatusBadRequest
				http.Error(w, fmt.Sprintf("%s already exists and is a directory, please choose another path\n", fpmd), ctx.statusCode)
				return
			} else {
				ctx.path = fpmd
				err = ctx.View(param_version)
			}
		} else { // both fp and fpmd does not exists
			ctx.path = fpmd
			err = ctx.Edit(param_version)
		}
	} else {
		// method not allowed
		ctx.statusCode = http.StatusBadRequest
		http.Error(w, r.Method+" method not allowed for view", ctx.statusCode)
		return
	}

	if err != nil {
		if ctx.statusCode == http.StatusOK {
			ctx.statusCode = http.StatusBadRequest
		}
		http.Error(w, err.Error(), ctx.statusCode)
	}
}

func main() {
	parseConfig()
	bootstrap()

	// try open the repo
	repo, err := git.OpenRepository(".")
	if err != nil {
		log.Printf("git repository not found at current directory. please use `-init` switch or run `git init` in this directory")
		log.Fatal(err)
		os.Exit(2)
	} else {
		repo.Free()
	}

	// load auth file
	if _, err := os.Stat(wikiConfig.auth); len(wikiConfig.auth) > 0 && (!os.IsNotExist(err)) {
		authenticator = auth.NewBasicAuthenticator("strapdown.ztx.io", auth.HtpasswdFileProvider(wikiConfig.auth)) // should we replace the url here?
		log.Printf("use authentication file: %s", wikiConfig.auth)
	} else {
		log.Printf("authentication file not exist, disable http authentication")
	}

	if _, err := os.Stat(".md"); os.IsNotExist(err) {
		// release a default .md
		log.Print("Release default .md")

		file, err := Asset("_static/.md")
		if err != nil {
			log.Printf("[ WARN ] fail to load .md")
		}
		err = ioutil.WriteFile(".md", file, 0644)
		if err != nil {
			log.Printf("[ WARN ] cannot write default .md: %v", err)
		}
	}

	if _, err := os.Stat("favicon.ico"); os.IsNotExist(err) {
		// release the files
		log.Print("Release the favicon.ico")

		file, err := Asset("_static/fav.ico")
		if err != nil {
			log.Printf("[ WARN ] fail to load favicon.ico")
		}
		err = ioutil.WriteFile("favicon.ico", file, 0644)
		if err != nil {
			log.Printf("[ WARN ] cannot write default favicon.ico: %v", err)
		}
	}

	http.HandleFunc("/", handleFunc)

	// listen on the (multi) addresss
	cnt := 0
	ch := make(chan bool)
	for _, host := range strings.Split(wikiConfig.addr, ",") {
		cnt += 1
		log.Printf("[ %d ] listening on %s", cnt, host)
		go func(h string, aid int) {
			e := http.ListenAndServe(h, nil)
			if e != nil {
				log.Printf("[ %d ] failed to bind on %s: %v", aid, h, e)
				ch <- false
			} else {
				ch <- true
			}
		}(host, cnt)
	}

	for cnt > 0 {
		<-ch
		cnt -= 1
	}
}
