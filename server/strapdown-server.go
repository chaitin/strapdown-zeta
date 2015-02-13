package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/libgit2/git2go"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var port = flag.Int("port", 8080, "The port for the server to listen")
var addr = flag.String("address", "0.0.0.0", "Listening address")
var initgit = flag.Bool("init", false, "init git repository before running, just like `git init`")
var root = flag.String("dir", "", "The root directory for the git/wiki")

type Config struct {
	Title         string
	Theme         string
	Toc           bool
	HeadingNumber string
	Host          string
	Content       template.HTML
}

func (config *Config) FillDefault(content []byte) {
	if config.Title == "" {
		config.Title = "Wiki"
	}
	if config.Theme == "" {
		config.Theme = "cerulean"
	}
	if config.HeadingNumber == "" {
		config.HeadingNumber = "false"
	}
	if config.Host == "" {
		config.Host = "cdn.ztx.io"
	}
	if config.Content == "" {
		config.Content = template.HTML(content)
	}
}

var viewTemplate, editTemplate *template.Template

func init_after_main() { // init after main because we need to chdir first, then write the default favicon
	var err error
	viewTemplate, err = template.New("view").Parse("<!DOCTYPE html> <html> <title>{{.Title}}</title> <meta charset=\"utf-8\"> <xmp theme=\"{{.Theme}}\" toc=\"{{.Toc}}\" heading_number=\"{{.HeadingNumber}}\" style=\"display:none;\">\n{{.Content}}\n</xmp> <script src=\"http://{{.Host}}/strapdown/strapdown.min.js\"></script> </html>\n")
	if err != nil {
		log.Fatalf("cannot parse view template")
	}
	editTemplate, err = template.New("edit").Parse("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge,chrome=1\"><title>{{.Title}}</title><link rel=\"stylesheet\" href=\"http://{{.Host}}/strapdown/themes/cerulean.min.css\" /><link rel=\"stylesheet\" href=\"http://{{.Host}}/strapdown/themes/bootstrap-responsive.min.css\" /><style type=\"text/css\" media=\"screen\">html, body {height: 100%;overflow: hidden;margin: 0;padding: 0;}#editor {margin: 0;position: absolute;top: 51px;bottom: 0;left: 0;right: 0;}@media (max-width: 980px) {#editor {top: 60px;}}</style></head><body><div class=\"navbar navbar-fixed-top\"><div class=\"navbar-inner\"><div style=\"padding:0 20px\"><a class=\"btn btn-navbar\" data-toggle=\"collapse\" data-target=\".navbar-responsive-collapse\"><span class=\"icon-bar\"></span><span class=\"icon-bar\"></span><span class=\"icon-bar\"></span></a><div id=\"headline\" class=\"brand\"> {{.Title}} </div><div class=\"nav-collapse collapse navbar-responsive-collapse pull-right\"> <form class=\"nav\" method=\"POST\" name=\"body\"><input id=\"savValue\" type=\"hidden\" name=\"body\" value=\"\" /><button class=\"btn btn-default btn-sm\" type=\"submit\">Save</button></form></div></div> </div></div><xmp id=\"editor\">{{.Content}}</xmp><script src=\"http://{{.Host}}/ace/ace.js\" type=\"text/javascript\" charset=\"utf-8\"></script><script src=\"http://{{.Host}}/strapdown/edit.js\" type=\"text/javascript\" charset=\"utf-8\"></script></body></html>\n")
	if err != nil {
		log.Fatalf("cannot parse edit template")
	}
	defaultFavicon, err := base64.StdEncoding.DecodeString("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")
	if err != nil {
		log.Printf("[ WARN ] %v", err)
		return
	}
	if _, err := os.Stat("favicon.ico"); os.IsNotExist(err) {
		log.Printf("write default favicon.ico to working directory")
		err = ioutil.WriteFile("favicon.ico", defaultFavicon, 0644)
		if err != nil {
			log.Printf("[ WARN ] cannot write default favicon.ico: %v", err)
		}
	}
}

func save_and_commit(fp string, content []byte, comment string, author string) error {
	var err error

	err = os.MkdirAll(path.Dir(fp), 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fp, content, 0600)
	if err != nil {
		return err
	}

	repo, err := git.OpenRepository(".")
	if err != nil {
		return err
	}
	index, err := repo.Index()
	if err != nil {
		return err
	}
	err = index.AddByPath(fp)
	if err != nil {
		return err
	}

	treeId, err := index.WriteTree()
	if err != nil {
		return err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		return err
	}

	sig := &git.Signature{
		Name:  author,
		Email: "strapdown@gmail.com",
		When:  time.Now(),
	}

	currentBranch, err := repo.Head()
	if err == nil && currentBranch != nil {
		currentTip, err2 := repo.LookupCommit(currentBranch.Target())
		if err2 != nil {
			return err2
		}
		_, err = repo.CreateCommit("HEAD", sig, sig, comment, tree, currentTip)
	} else {
		_, err = repo.CreateCommit("HEAD", sig, sig, comment, tree)
	}

	if err != nil {
		return err
	}
	return nil
}

func remote_ip(r *http.Request) string {
	ret := r.RemoteAddr
	i := strings.IndexByte(ret, ':')
	if i > -1 {
		ret = ret[:i]
	}
	if r.Header.Get("X-FORWARDED-FOR") != "" {
		if strings.Index(ret, "127.0.0.1") == 0 {
			ret = r.Header.Get("X-FORWARDED-FOR")
		} else {
			ret = fmt.Sprintf("%s,%s", ret, r.Header.Get("X-FORWARDED-FOR"))
		}
	}
	return ret
}

func getFile(repo *git.Repository, commit *git.Commit, fileName string) (*string, error) {
	var err error
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	enter := tree.EntryByName(fileName)
	if enter == nil {
		return nil, err
	}

	oid := enter.Id
	blb, err := repo.LookupBlob(oid)
	if err != nil {
		return nil, err
	}

	ret := string(blb.Contents())
	return &ret, nil
}

func getFileOfVersion(fileName string, version string) ([]byte, error) {
	var err error

	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}

	currentBranch, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.LookupCommit(currentBranch.Target())
	if err != nil {
		return nil, err
	}

	vl := len(version)

	if vl < 4 || vl > 40 {
		return nil, fmt.Errorf("version length should be in range [4, 40], provided %d", vl)
	}

	for commit != nil {
		if commit.Id().String()[0:len(version)] == version {
			str, err := getFile(repo, commit, fileName)
			if err != nil {
				return nil, err
			}

			var s []byte
			if str != nil {
				s = []byte(*str)
			}
			return s, nil
		}
		commit = commit.Parent(0)
	}
	return nil, nil
}

// copied from http://golang.org/src/net/http/fs.go
func safe_open(base string, name string) (*os.File, error) {
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

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&#34;",
	"'", "&#39;",
)

func handle(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	defer func() {
		log.Printf("[ %s ] - %d %s", r.Method, statusCode, r.URL.String())
	}()

	var err error

	q := r.URL.Query()

	_, doedit := q["edit"]
	version, doversion := q["version"]

	fp := r.URL.Path[1:]

	if strings.HasPrefix(fp, ".git/") || fp == ".git" {
		statusCode = http.StatusForbidden
		http.Error(w, "access of .git directory not allowed", statusCode)
		return
	}

	if stat, err := os.Stat(fp); err == nil {
		if !stat.IsDir() {
			http.ServeFile(w, r, fp)
			return
		} else {
			if statmd, err := os.Stat(fp + ".md"); err == nil && !statmd.IsDir() {
				// if the following cases, dont list dir:
				// if /path/to/dir/.md exists, just show its content instead of listing dir
				// if doedit, goto edit mode
				// if it's root directory /, goto edit mode or view mode
			} else if !doedit && len(fp) > 0 {
				// list dir here

				dirfile, err := safe_open(fp, "")
				if err != nil {
					statusCode = http.StatusBadRequest
					http.Error(w, err.Error(), statusCode)
					return
				}

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				// w.Write(view_head)
				fmt.Fprintf(w, "# Directory listing for %s\n", fp)
				for {
					dirs, err := dirfile.Readdir(100)
					if err != nil || len(dirs) == 0 {
						break
					}
					for _, d := range dirs {
						name := d.Name()
						if d.IsDir() {
							name += "/"
						}
						dirurl := url.URL{Path: path.Join("/", fp, name)}
						fmt.Fprintf(w, " - [%s](%s)\n", htmlReplacer.Replace(name), dirurl.String())
					}
				}
				// w.Write(view_tail)
				return
			}
		}
	}

	fp = r.URL.Path[1:] + ".md"

	if r.Method == "POST" || r.Method == "PUT" {
		err := save_and_commit(fp, []byte(r.FormValue("body")), "update "+fp, "anonymous@"+remote_ip(r))
		if err != nil {
			statusCode = http.StatusInternalServerError
			http.Error(w, err.Error(), statusCode)
			return
		}
		statusCode = http.StatusFound
		http.Redirect(w, r, r.URL.Path, statusCode)
		return
	}

	var content []byte

	handleEdit := func() {
		custom_option, err := ioutil.ReadFile(fp + ".option.json")
		var config Config = Config{}
		if err == nil {
			json.Unmarshal(custom_option, &config)
		}
		config.FillDefault(content)
		editTemplate.Execute(w, config)
	}

	if doversion && len(version) > 0 && len(version[0]) > 0 {
		content, err = getFileOfVersion(fp, version[0])
		if err != nil {
			statusCode = http.StatusBadRequest
			http.Error(w, err.Error(), statusCode)
			return
		}
		if content == nil {
			statusCode = http.StatusNotFound
			http.Error(w, "Error : Can not find "+fp+" of version "+version[0], statusCode)
			return
		}
	} else {
		doversion = false
		content, err = ioutil.ReadFile(fp)

		if err != nil {
			if _, err := os.Stat(fp); err != nil {
				// file not exist or permission denied, enter edit mode
				handleEdit()
			} else {
				statusCode = http.StatusNotFound
				http.Error(w, err.Error(), statusCode)
			}
			return
		}
	}

	if doedit {
		// enter edit mode
		handleEdit()
		return
	}

	custom_view_head, errh := ioutil.ReadFile(fp + ".head")
	custom_view_tail, errt := ioutil.ReadFile(fp + ".tail")
	if errh == nil && errt == nil {
		w.Write(custom_view_head)
		w.Write(content)
		w.Write(custom_view_tail)
	} else {
		custom_view_option, errv := ioutil.ReadFile(fp + ".option.json")
		var config Config = Config{}
		if errv == nil {
			json.Unmarshal(custom_view_option, &config)
		}
		config.FillDefault(content)
		viewTemplate.Execute(w, config)
	}
}

func main() {
	flag.Parse()
	var err error

	if len(*root) > 0 {
		err = os.Chdir(*root)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Printf("chdir to %s", *root)
	}

	if *initgit {
		if _, err = git.OpenRepository("."); err != nil {
			_, err = git.InitRepository(".", false)
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Printf("git init finished at .")
		} else {
			log.Printf("git repository already found, skip git init")
		}
	}
	_, err = git.OpenRepository(".")
	if err != nil {
		log.Printf("git repository not found at current directory. please use `-init` switch or run `git init` in this directory")
		log.Fatal(err)
		return
	}
	init_after_main()
	http.HandleFunc("/", handle)
	host := fmt.Sprintf("%s:%d", *addr, *port)
	log.Printf("listening on %s", host)
	l, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{}
	s.Serve(l)
}
