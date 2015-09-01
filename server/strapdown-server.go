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
	toc            string
	verbose        bool
	version        bool
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
	isFolder   bool
	hasFile    bool
	username   string
	statusCode int
}

var wikiConfig Config // the global config file
var templates map[string]*template.Template
var authenticator *auth.BasicAuth

const VERSION = "0.4.1" // major.minor.patch
const POWERED_BY = "Strapdown Server (v" + VERSION + ")"

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

func (this *RequestContext) parseInfo() {
	fp := this.req.URL.Path
	if this.req.Method == "GET" {
		if fp[len(fp)-1] != '/' && strings.Contains(path.Base(fp), ".") {
			//suffixed file
			data, err := os.Stat(fp[1:])
			this.hasFile = (err == nil)
			if this.hasFile {
				this.isMarkdown = false
				this.isFolder = data.IsDir()
			} else {
				this.isMarkdown = true
				fp += ".md"
				data, err := os.Stat(fp[1:])
				if this.hasFile = (err == nil); this.hasFile {
					this.isFolder = data.IsDir()
				} else {
					this.isFolder = false
				}
			}
			// we want the page show the original .md if we have xxx.md in the url
			// and treat any other unnormal urls as the the markdown, so we can edit it, eg /xxx.dsd/ss.dd style url
		} else {
			// for the urls with no .md and not existed, regards as markdown file and edit
			_, err := os.Stat(fp[1:])
			this.hasFile = (err == nil)
			log.Print(this.hasFile)
			if this.hasFile {
				this.isMarkdown = false
				this.isFolder = false
			} else {
				fp += ".md"
				this.isMarkdown = true
				data, err := os.Stat(fp[1:])
				this.hasFile = (err == nil)
				if this.hasFile {
					this.isFolder = data.IsDir()
				} else {
					this.isFolder = false
				}
			}
		}
	} else {
		this.isMarkdown = false
		this.isFolder = false
	}

	this.path = fp[1:]

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
	w := *this.res

	// disable log for static
	if wikiConfig.verbose && !strings.HasPrefix(this.path, "_static") && !strings.HasPrefix(this.path, "favicon.ico") {
		log.Printf("[ DEBUG ] Path <%s> Markdown: %t Folder: %t Existed: %t", this.path, this.isMarkdown, this.isFolder, this.hasFile)
	}

	version_ary, hasversion := q["version"]
	this.Version = getVersion(hasversion, version_ary)
	if this.req.Method == "GET" {
		if this.isFolder && this.hasFile {
			http.Redirect(w, req, this.path+"/", http.StatusTemporaryRedirect)
			return nil
		}
		if this.isMarkdown {
			if this.hasFile {
				data, err := ioutil.ReadFile(this.path)
				if err == nil {
					this.Content = template.HTML(string(data))
				}
			}
			if _, history := q["history"]; history {
				return this.History()
			} else if _, edit := q["edit"]; edit {
				return this.Edit()
			} else if diff_ary, diff := q["diff"]; diff {
				return this.Diff(diff_ary)
			} else {
				if this.hasFile {
					return this.View()
				} else {
					file := path.Base(this.path)
					folder := path.Dir(this.path)
					_, err := os.Stat(folder)

					if err == nil && file == ".md" && !this.hasFile && folder != "." {
						this.path = folder
						return this.Listdir()
					}
					if this.hasFile {
						return this.View()
					} else {
						return this.Edit()
					}
				}
			}
		} else {
			if this.hasFile {
				file, err := GetFileOfVersion(this.path, this.Version)
				// when the file is not in the git commit, the file would be []
				// we should treat this as fail
				if err == nil && len(file) != 0 {
					w.Write(file)
					return nil
				} else {
					var mimetype string = "application/octet-stream"
					lastdot := strings.LastIndex(this.path, ".")
					if lastdot > -1 {
						mimetype = mime.TypeByExtension(this.path[lastdot:])
					}
					w.Header().Set("Content-Type", mimetype)

					data, err := ioutil.ReadFile(this.path)
					w.Write(data)
					return err
				}
			} else {
				http.NotFound(*this.res, this.req)
				this.statusCode = http.StatusNotFound
				return nil
			}
		}
	} else if this.req.Method == "POST" || this.req.Method == "PUT" {
		if strings.HasSuffix(this.path, "option.json") {
			this.statusCode = http.StatusForbidden
			return errors.New("Uploading option.json is not allowed")
		}
		return this.Update()
	}
	return nil
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	// cache is evil

	var ctx RequestContext
	ctx.req = r
	ctx.res = &w
	ctx.statusCode = http.StatusOK
	ctx.Title = wikiConfig.title
	ctx.Theme = wikiConfig.theme
	ctx.Toc = wikiConfig.toc
	ctx.HeadingNumber = wikiConfig.heading_number
	ctx.Host = wikiConfig.host

	w.Header().Set("X-Powered-By", POWERED_BY)

	defer func() {
		log.Printf("[ %s ] - %d %s", r.Method, ctx.statusCode, r.URL.String())
	}()

	if authenticator != nil { // check http auth
		if ctx.username = authenticator.CheckAuth(r); ctx.username == "" {
			authenticator.RequireAuth(w, r)
			return
		}
	}

	ctx.parseInfo()

	if strings.HasSuffix(ctx.path, "_static") || strings.HasSuffix(ctx.path, "favicon.ico") {
		w.Header().Set("Cache-Control", "max-age=86400, public")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, post-check=0, pre-check=0, max-age=0")
		w.Header().Set("Expires", "Sun, 19 Nov 1978 05:00:00 GMT")
	}

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

	if wikiConfig.version {
		fmt.Printf("Strapdown Wiki Server - v%s\n", VERSION)
		os.Exit(0)
	}

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

	if _, err := os.Stat("_static"); os.IsNotExist(err) {
		// release the files
		log.Print("Seems you don't have `_static` folder, release one to hold the static file")
		files := AssetNames()

		for _, name := range files {
			if strings.HasSuffix(name, ".html") || strings.HasSuffix(name, "fav.ico") {
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
			if err != nil {
				log.Printf("[ WARN ] cannot write file: %v", err)
			}
		}
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
