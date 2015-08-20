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
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
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
	isMarkdown bool
	hasFile    bool
	username   string
	statusCode int
}

var wikiConfig Config // the global config file
var templates map[string]*template.Template
var authenticator *auth.BasicAuth

func parseConfig() {
	flag.StringVar(&wikiConfig.addr, "addr", ":8080", "Listening `host:port`, you can specify multiple listening address separated by comma, e.g. (127.0.0.1:8080,192.168.1.2:8080)")
	flag.BoolVar(&wikiConfig.init, "init", false, "init git repository before running, just like `git init`")
	flag.StringVar(&wikiConfig.root, "dir", "", "The root directory for the git/wiki")
	flag.StringVar(&wikiConfig.auth, "auth", ".htpasswd", "Default auth file to use as authentication, authentication will be disabled if auth file not exist")
	flag.StringVar(&wikiConfig.host, "host", "cdn.ztx.io", "Default host hosting the strapdown static files")
	flag.StringVar(&wikiConfig.heading_number, "heading_number", "false", "set default value for showing heading number")
	flag.StringVar(&wikiConfig.title, "title", "Wiki", "default title for wiki pages")
	flag.StringVar(&wikiConfig.theme, "theme", "cerulean", "default theme for strapdown")
	flag.IntVar(&wikiConfig.histsize, "histsize", 30, "default history size")
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

func getVersion(doversion bool, version_ary []string) string {
	if doversion {
		if len(version_ary) > 0 && len(version_ary[0]) > 0 {
			return version_ary[0]
		}
	}
	repo, err := git.OpenRepository(".")
	if err != nil {
		return ""
	}
	head, err := repo.Head()
	if err != nil {
		return ""
	}
	repo.Free()
	return head.Target().String()
}

func bootstrap() {

	templates = make(map[string]*template.Template)

	pages := []string{"view", "listdir", "history", "diff", "edit"}
	for _, element := range pages {
		data, err := Asset("_static/" + element + ".html")
		if err != nil {
			log.Fatalf("fail to load the %s.html", element)
		}
		templates[element], err = template.New(element).Parse(string(data))
		if err != nil {
			log.Fatalf("cannot parse %s template", element)
		}
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
}

// todo: rename
func (this *RequestContext) parsePath() {
	fp := this.req.URL.Path[1:]
	if strings.Contains(fp, ".") {
		//suffixed file
		this.isMarkdown = strings.HasSuffix(fp, ".md")
	} else {
		fp += ".md"
		this.isMarkdown = true
	}
	this.path = fp

	_, err := os.Stat(fp)
	this.hasFile = err == nil

	i := strings.IndexByte(this.req.RemoteAddr, ':')
	if i > -1 {
		this.ip = this.req.RemoteAddr[:i]
	}
	if this.req.Header.Get("X-FORWARDED-FOR") != "" {
		if strings.Index(this.ip, "127.0.0.1") == 0 {
			this.ip = this.req.Header.Get("X-FORWARDED-FOR")
		} else {
			this.ip = fmt.Sprintf("%s,%s", this.ip, this.req.Header.Get("X-FORWARDED-FOR"))
		}
	}
}

func (this *RequestContext) parseAndDo(req *http.Request) error {
	q := req.URL.Query()

	version_ary, hasversion := q["version"]
	this.Version = getVersion(hasversion, version_ary)

	if this.req.Method == "GET" {
		if this.isMarkdown {
			if this.hasFile {
				this.Content = template.HTML(string(Read(this.path)))
			}
			if _, history := q["history"]; history {
				return this.History()
			} else if _, edit := q["edit"]; edit {
				return this.Edit()
			} else if diff_ary, diff := q["diff"]; diff {
				return this.Diff(diff_ary)
			} else {
				if strings.HasSuffix("/"+this.path, "/.md") {
					return this.Listdir()
				}
				if this.hasFile {
					return this.View()
				} else {
					return this.Edit()
				}
			}
		} else {
			if this.hasFile {
				var w = *this.res
				file, err := GetFileOfVersion(this.path, this.Version)
				if err == nil {
					w.Write(file)
					return nil
				} else {
					//
					file := Read(this.path)
					w.Write(file)
					return nil
				}
			} else {
				http.NotFound(*this.res, this.req)
				this.statusCode = http.StatusNotFound
				return nil
			}
		}
	} else if this.req.Method == "POST" || this.req.Method == "PUT" {
		return this.Update()
	}
	return nil
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	// cache is evil
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, post-check=0, pre-check=0, max-age=0")
	w.Header().Set("Expires", "Sun, 19 Nov 1978 05:00:00 GMT")
	var ctx RequestContext
	ctx.req = r
	ctx.res = &w
	ctx.statusCode = http.StatusOK
	ctx.Title = wikiConfig.title
	ctx.Theme = wikiConfig.theme
	ctx.Toc = false
	ctx.HeadingNumber = wikiConfig.heading_number
	ctx.Host = wikiConfig.host

	defer func() {
		log.Printf("[ %s ] - %d %s", r.Method, ctx.statusCode, r.URL.String())
	}()

	if authenticator != nil { // check http auth
		if ctx.username = authenticator.CheckAuth(r); ctx.username == "" {
			authenticator.RequireAuth(w, r)
			return
		}
	}

	ctx.parsePath()

	// forbidden any access of git related object
	if strings.HasPrefix(ctx.path, ".git/") || ctx.path == ".git" || ctx.path == ".gitignore" || ctx.path == ".gitmodules" {
		ctx.statusCode = http.StatusForbidden
		http.Error(w, "access of .git related files/directory not allowed", ctx.statusCode)
		return
	}
	if len(wikiConfig.auth) > 0 && ctx.path == wikiConfig.auth || ctx.path == wikiConfig.auth {
		ctx.statusCode = http.StatusForbidden
		http.Error(w, "access of password file not allowed", ctx.statusCode)
		return
	}

	err := ctx.parseAndDo(r)
	if err != nil {
		http.Error(w, err.Error(), ctx.statusCode)
		log.Printf("Failed: %v", err)
		return
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

	if _, err := os.Stat("favicon.ico"); os.IsNotExist(err) {
		ico, err := Asset("_static/fav.ico")
		err = ioutil.WriteFile("favicon.ico", ico, 0644)
		if err != nil {
			log.Printf("[ WARN ] cannot write default favicon.ico: %v", err)
		} else {
			log.Printf("[ ctx ] not found favicon.ico, write a default one")
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
