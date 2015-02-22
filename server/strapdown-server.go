package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/libgit2/git2go"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var addr = flag.String("addr", ":8080", "Listening `host:port`, you can specify multiple listening address separated by comma, e.g. (127.0.0.1:8080,192.168.1.2:8080)")
var initgit = flag.Bool("init", false, "init git repository before running, just like `git init`")
var root = flag.String("dir", "", "The root directory for the git/wiki")
var default_host = flag.String("host", "cdn.ztx.io", "Default host hosting the strapdown static files")
var default_heading_number = flag.String("heading_number", "false", "set default value for showing heading number")
var default_title = flag.String("title", "Wiki", "default title for wiki pages")
var default_theme = flag.String("theme", "cerulean", "default theme for strapdown")
var default_histsize = flag.Int("histsize", 30, "default history size")

type DirEntry struct {
	Urlpath string
	Name    string
	Size    int64
	IsDir   bool
	ModTime time.Time
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

type CommitEntry struct {
	Id        string
	Timestamp time.Time
	Author    string
	Message   string
}

func (this *CommitEntry) ShortHash() string {
	return this.Id[:11]
}

type Config struct {
	Title         string
	Theme         string
	Toc           bool
	HeadingNumber string
	Host          string
	Content       template.HTML
	DirEntries    []DirEntry
	CommitEntries []CommitEntry
}

func (config *Config) FillDefault(content []byte) {
	if config.Title == "" {
		config.Title = *default_title
	}
	if config.Theme == "" {
		config.Theme = *default_theme
	}
	if config.HeadingNumber == "" {
		config.HeadingNumber = *default_heading_number
	}
	if config.Host == "" {
		config.Host = *default_host
	}
	if config.Content == "" {
		config.Content = template.HTML(content)
	}
}

var viewTemplate, editTemplate, listdirTemplate, historyTemplate *template.Template

func init_after_main() { // init after main because we need to chdir first, then write the default favicon
	var err error
	viewTemplate, err = template.New("view").Parse("<!DOCTYPE html> <html> <title>{{.Title}}</title> <meta charset=\"utf-8\"> <xmp theme=\"{{.Theme}}\" toc=\"{{.Toc}}\" heading_number=\"{{.HeadingNumber}}\" style=\"display:none;\">\n{{.Content}}\n</xmp> <script src=\"http://{{.Host}}/strapdown/strapdown.min.js\"></script> </html>\n")
	if err != nil {
		log.Fatalf("cannot parse view template")
	}
	editTemplate, err = template.New("edit").Parse("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge,chrome=1\"><title>{{.Title}}</title><link rel=\"stylesheet\" href=\"http://{{.Host}}/strapdown/themes/cerulean.min.css\" /><style type=\"text/css\" media=\"screen\">html, body {height: 100%;overflow: hidden;margin: 0;padding: 0;}#editor {margin: 0;position: absolute;top: 51px;bottom: 0;left: 0;right: 0;}</style></head><body><div class=\"navbar navbar-fixed-top\"><div class=\"navbar-inner\"><div style=\"padding:0 20px\"><a class=\"btn btn-navbar\" data-toggle=\"collapse\" data-target=\".navbar-responsive-collapse\"><span class=\"icon-bar\"></span><span class=\"icon-bar\"></span><span class=\"icon-bar\"></span></a><div id=\"headline\" class=\"brand\"> {{.Title}} </div><div class=\"nav-collapse collapse navbar-responsive-collapse pull-right\"> <form class=\"nav\" method=\"POST\" action=\"?edit\" name=\"body\"><input id=\"savValue\" type=\"hidden\" name=\"body\" value=\"\" /><button class=\"btn btn-default btn-sm\" type=\"submit\">Save</button></form></div></div> </div></div><xmp id=\"editor\">{{.Content}}</xmp><script src=\"http://{{.Host}}/ace/ace.js\" type=\"text/javascript\" charset=\"utf-8\"></script><script src=\"http://{{.Host}}/strapdown/edit.js\" type=\"text/javascript\" charset=\"utf-8\"></script></body></html>\n")
	if err != nil {
		log.Fatalf("cannot parse edit template")
	}
	listdirTemplate, err = template.New("listdir").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="http://{{.Host}}/strapdown/themes/cerulean.min.css" />
  <link rel="stylesheet" href="http://{{.Host}}/strapdown/themes/bootstrap-responsive.min.css" />
  <style type="text/css" media="screen">
    #list {
        margin: 51px auto;
        -webkit-box-sizing: border-box; /* Safari, other WebKit */
        -moz-box-sizing: border-box;    /* Firefox, other Gecko */
        box-sizing: border-box;         /* Opera/IE 8+ */
    }
    #list table {
        -webkit-box-sizing: border-box; /* Safari, other WebKit */
        -moz-box-sizing: border-box;    /* Firefox, other Gecko */
        box-sizing: border-box;         /* Opera/IE 8+ */
        max-width: 100%;
        border-collapse: collapse;
        word-wrap: break-word;
        word-break: break-all;
    }
    #list td, #list th {
        display: table-cell;
        font-size: 16px;
        height: 26px;
        line-height: 26px;
        text-align: center;
        vertical-align: top;
        min-width: 100px;
        word-wrap: break-word;
        word-break: break-all;
    }
    #list tr>th:nth-child(1) {
        text-align: left;
    }
    #list tr>td:nth-child(1) {
        text-align: left;
        width: auto;
        white-space: normal;
        max-width: 90%;
        text-align:left;
    }
    #list tr>td>a {
        display: block;
        color: #333;
        text-decoration: none;
    }
    #list .endslash {
        color: #6299fe;
        font-weight: bold;
    }
    @media (max-width: 980px) {
      #list {
        margin: 0 10px 10px 0;
        -webkit-box-sizing: border-box; /* Safari, other WebKit */
        -moz-box-sizing: border-box;    /* Firefox, other Gecko */
        box-sizing: border-box;         /* Opera/IE 8+ */
      }
    }
  </style>
</head>
<body>
  <div class="navbar navbar-fixed-top">
    <div class="navbar-inner">
      <div class="container">
        <div id="headline" class="brand"> Directory Listing of {{.Title}} </div>
      </div> 
    </div>
  </div>
  <div id="list" class="container">
    <hr />
    <table class="table table-hover">
      <thead>
        <tr>
          <th>Filename</th>
          <th>Size</th>
          <th>Datetime</th>
        </tr>
      </thead>
      <tbody>
        {{ range $index, $element := .DirEntries }}
        <tr>
          <td><a href="{{$element.Urlpath}}">{{$element.Name}} {{ if $element.IsDir }} <span class="endslash">/</span> {{ end }} </a></td>
          <td><a href="{{$element.Urlpath}}" title="{{$element.Size}}B">{{$element.ReadableSize true}}</a></td>
          <td><a href="{{$element.Urlpath}}">{{$element.ModTime.Format "2006-01-02 15:04:05"}}</a></td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    <hr />
  </div>
</body>
</html>
`)
	if err != nil {
		log.Fatalf("cannot parse listdir template: %v", err)
	}
	historyTemplate, err = template.New("history").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="http://{{.Host}}/strapdown/themes/cerulean.min.css" />
  <link rel="stylesheet" href="http://{{.Host}}/strapdown/themes/bootstrap-responsive.min.css" />
  <style type="text/css" media="screen">
    #list {
        margin: 51px auto;
        -webkit-box-sizing: border-box; /* Safari, other WebKit */
        -moz-box-sizing: border-box;    /* Firefox, other Gecko */
        box-sizing: border-box;         /* Opera/IE 8+ */
    }
    #list .table td input {
        padding: 0 4px;
        margin: 0px 32px 0 12px;
        vertical-align: middle;
        width: 12px;
        height: 12px;
    }
    #list .table td span {
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 650px;
        display: inline-block;
    }
    #list .table td input+a {
        display: inline-block;
    }
    #list .table td input+a:hover {
        text-decoration: underline;
    }
    #list table {
        -webkit-box-sizing: border-box; /* Safari, other WebKit */
        -moz-box-sizing: border-box;    /* Firefox, other Gecko */
        box-sizing: border-box;         /* Opera/IE 8+ */
        max-width: 100%;
        border-collapse: collapse;
        word-wrap: break-word;
        word-break: break-all;
    }
    #list td, #list th {
        display: table-cell;
        font-size: 16px;
        height: 26px;
        line-height: 26px;
        text-align: center;
        vertical-align: top;
        min-width: 100px;
        word-wrap: break-word;
        word-break: break-all;
    }
    #list tr>th:nth-child(1) {
        text-align: left;
        padding-left: 64px;
    }
    #list tr>td:nth-child(1) {
        text-align: left;
        font-family: Courier;
    }
    #list tr>th:nth-child(2) {
        text-align: left;
    }
    #list tr>td:nth-child(2) {
        text-align: left;
        width: auto;
        text-align:left;
    }
    #list tr>td>a {
        display: block;
        color: #333;
        text-decoration: none;
    }
    #list .endslash {
        color: #6299fe;
        font-weight: bold;
    }
    @media (max-width: 980px) {
      #list {
        margin: 0 10px 10px 0;
        -webkit-box-sizing: border-box; /* Safari, other WebKit */
        -moz-box-sizing: border-box;    /* Firefox, other Gecko */
        box-sizing: border-box;         /* Opera/IE 8+ */
      }
    }
  </style>
</head>
<body>
  <div class="navbar navbar-fixed-top">
    <div class="navbar-inner">
      <div class="container">
        <div id="headline" class="brand"> History of {{.Title}} </div>
      </div> 
    </div>
  </div>
  <div id="list" class="container">
    <hr />
    <table class="table table-hover">
      <thead>
        <tr>
          <th>Revision</th>
          <th>Comment</th>
          <th>Timestamp</th>
          <th>Author</th>
        </tr>
      </thead>
      <tbody>
        {{ range $index, $element := .CommitEntries }}
        <tr>
          <td><input type="checkbox" /><a href="?version={{$element.Id}}">{{ $element.ShortHash }}</a></td>
          <td><span>{{$element.Message}}</span></td>
          <td>{{$element.Timestamp.Format "2006-01-02 15:04:05"}}</td>
          <td>{{$element.Author}}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    <hr />
  </div>
</body>
</html>
`)
	if err != nil {
		log.Fatalf("cannot parse history template: %v", err)
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
	defer repo.Free()

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
	defer tree.Free()

	var entry *git.TreeEntry
	if strings.IndexByte(fileName, '/') >= 0 {
		entry, err = tree.EntryByPath(fileName)
	} else {
		entry = tree.EntryByName(fileName)
		err = nil
	}
	if entry == nil || err != nil {
		return nil, err
	}

	oid := entry.Id
	blb, err := repo.LookupBlob(oid)
	if err != nil {
		return nil, err
	}
	defer blb.Free()

	ret := string(blb.Contents())
	return &ret, nil
}

func getFileOfVersion(fileName string, version string) ([]byte, error) {
	var err error

	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}
	defer repo.Free()

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

func history(fp string, size int) ([]CommitEntry, error) {
	if len(fp) == 0 {
		return nil, nil
	}
	var err error
	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	revwalk, err := repo.Walk()
	if err != nil {
		return nil, err
	}
	defer revwalk.Free()

	err = revwalk.PushHead()
	if err != nil {
		return nil, err
	}

	revwalk.Sorting(git.SortTime)

	var filehistory []CommitEntry
	cnt := 0

	err = revwalk.Iterate(func(commit *git.Commit) bool {
		defer commit.Free()

		tree, err := commit.Tree()
		if err != nil {
			return false
		}
		defer tree.Free()

		var entry *git.TreeEntry
		if strings.IndexByte(fp, '/') >= 0 {
			entry, err = tree.EntryByPath(fp)
		} else {
			entry = tree.EntryByName(fp)
			err = nil
		}

		if entry != nil && err == nil {
			filehistory = append(filehistory, CommitEntry{Id: commit.Id().String(), Message: commit.Message(), Author: commit.Author().Name, Timestamp: commit.Author().When})
			cnt += 1
			// log.Println(commit.Message(), commit.Id().String()[0:12])
			if size > 0 && cnt >= size {
				return false
			}
		}
		return true
	})

	return filehistory, nil
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

func handle(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	defer func() {
		log.Printf("[ %s ] - %d %s", r.Method, statusCode, r.URL.String())
	}()

	var err error

	fp := r.URL.Path[1:]
	fpmd := fp + ".md"

	// parse query and param first
	q := r.URL.Query()

	_, doedit := q["edit"]
	version_ary, doversion := q["version"]
	histsize_ary, dohistory := q["history"]

	var version string
	if doversion && len(version_ary) > 0 && len(version_ary[0]) > 0 {
		version = version_ary[0]
	} else {
		doversion = false
		version = ""
	}
	histsize := *default_histsize
	if dohistory && len(histsize_ary) > 0 {
		histsize, err = strconv.Atoi(histsize_ary[0])
		if err != nil {
			histsize = *default_histsize
		}
	}

	// forbidden any access of git related object
	if strings.HasPrefix(fp, ".git/") || fp == ".git" || fp == ".gitignore" || fp == ".gitmodules" {
		statusCode = http.StatusForbidden
		http.Error(w, "access of .git related files/directory not allowed", statusCode)
		return
	}

	// raw file, directory, markdown file all have a history in git, so handle them together here
	if dohistory {
		commit_history, err := history(fp, histsize)
		if err != nil || commit_history == nil || len(commit_history) == 0 {
			statusCode = http.StatusBadRequest
			if err != nil {
				http.Error(w, err.Error(), statusCode)
			} else {
				http.Error(w, "No commit history found for "+fp, statusCode)
			}
			return
		}
		custom_option, err := ioutil.ReadFile(fp + ".option.json")
		var config Config = Config{}
		if err == nil {
			json.Unmarshal(custom_option, &config)
		}
		if config.Title == "" {
			config.Title = fp
		}
		config.FillDefault(nil)
		config.CommitEntries = commit_history

		err = historyTemplate.Execute(w, config)
		if err != nil {
			log.Printf("[ ERR ] fill history template error: %v", err)
		}
		return
	}

	// handle post or put here, upload or edit or options
	if r.Method == "POST" || r.Method == "PUT" {
		// if dooptions { // handle options first
		// // TODO
		// }
		// edit or upload here
		var savefp string
		if doedit {
			savefp = fpmd
		} else {
			// do upload, or just raw post/put via command line(curl)
			savefp = fp
		}
		upload_content := []byte(r.FormValue("body"))
		if len(upload_content) == 0 && r.ContentLength > 0 {
			err = r.ParseMultipartForm(1048576 * 100)
			if err != nil {
				statusCode = http.StatusInternalServerError
				http.Error(w, err.Error(), statusCode)
				return
			}
			_, mh, err := r.FormFile("body")
			if err != nil {
				statusCode = http.StatusBadRequest
				http.Error(w, err.Error(), statusCode)
				return
			}
			buffer := &bytes.Buffer{}
			file, err := mh.Open()
			if err != nil {
				statusCode = http.StatusBadRequest
				http.Error(w, err.Error(), statusCode)
				return
			}
			defer file.Close()
			if _, err = io.Copy(buffer, file); err != nil {
				statusCode = http.StatusInternalServerError
				http.Error(w, err.Error(), statusCode)
				return
			}
			upload_content = buffer.Bytes()
		}
		err := save_and_commit(savefp, upload_content, "update "+savefp, "anonymous@"+remote_ip(r))
		if err != nil {
			statusCode = http.StatusInternalServerError
			http.Error(w, err.Error(), statusCode)
			return
		}
		statusCode = http.StatusFound
		http.Redirect(w, r, r.URL.Path, statusCode)
		return
	}

	// edit, view, with version considered below

	// check the raw file or directory first, no edit for raw file, no version for directory
	if fpstat, err := os.Stat(fp); err == nil {
		if !fpstat.IsDir() { // if the file exist, return the file with version handled
			if doversion {
				content, err := getFileOfVersion(fp, version)
				if err != nil {
					statusCode = http.StatusBadRequest
					http.Error(w, err.Error(), statusCode)
					return
				}
				if content == nil {
					statusCode = http.StatusNotFound
					http.Error(w, "Error : Can not find "+fp+" of version "+version, statusCode)
					return
				}
				w.Write(content)
			} else {
				http.ServeFile(w, r, fp)
			}
			return
		} else { // if it's a directory, then check .md first
			if statmd, err := os.Stat(fpmd); err == nil && !statmd.IsDir() {
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

				custom_option, err := ioutil.ReadFile(fp + ".option.json")
				var config Config = Config{}
				if err == nil {
					json.Unmarshal(custom_option, &config)
				}
				if config.Title == "" {
					config.Title = fp
				}
				config.FillDefault(nil)
				config.DirEntries = make([]DirEntry, 0, 16)

				fpurl := url.URL{Path: path.Join("/", fp, "..")}
				config.DirEntries = append(config.DirEntries, DirEntry{Name: "..", IsDir: true, Urlpath: fpurl.String(), Size: fpstat.Size(), ModTime: fpstat.ModTime()})

				for {
					dirs, err := dirfile.Readdir(128)
					if err != nil || len(dirs) == 0 {
						break
					}
					for _, d := range dirs {
						dirurl := url.URL{Path: path.Join("/", fp, d.Name())}
						dirurls := dirurl.String()
						if strings.HasSuffix(dirurls, ".md") {
							dirurls = strings.TrimSuffix(dirurls, ".md")
						}
						config.DirEntries = append(config.DirEntries, DirEntry{Name: d.Name(), IsDir: d.IsDir(), Urlpath: dirurls, Size: d.Size(), ModTime: d.ModTime()})
					}
				}
				err = listdirTemplate.Execute(w, config)
				if err != nil {
					log.Printf("[ ERR ] fill list dir template error: %v", err)
				}
				return
			}
		}
	}

	var content []byte

	handleEdit := func() {
		custom_option, err := ioutil.ReadFile(fpmd + ".option.json")
		var config Config = Config{}
		if err == nil {
			json.Unmarshal(custom_option, &config)
		}
		config.FillDefault(content)
		err = editTemplate.Execute(w, config)
		if err != nil {
			log.Printf("[ ERR ] fill edit template error: %v", err)
		}
	}

	if doversion {
		content, err = getFileOfVersion(fpmd, version)
		if err != nil {
			statusCode = http.StatusBadRequest
			http.Error(w, err.Error(), statusCode)
			return
		}
		if content == nil {
			statusCode = http.StatusNotFound
			http.Error(w, "Error : Can not find "+fpmd+" of version "+version, statusCode)
			return
		}
	} else {
		content, err = ioutil.ReadFile(fpmd)

		if err != nil {
			if _, err := os.Stat(fpmd); err != nil {
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

	custom_view_head, errh := ioutil.ReadFile(fpmd + ".head")
	custom_view_tail, errt := ioutil.ReadFile(fpmd + ".tail")
	if errh == nil && errt == nil {
		w.Write(custom_view_head)
		w.Write(content)
		w.Write(custom_view_tail)
	} else {
		custom_view_option, errv := ioutil.ReadFile(fpmd + ".option.json")
		var config Config = Config{}
		if errv == nil {
			json.Unmarshal(custom_view_option, &config)
		}
		config.FillDefault(content)
		err = viewTemplate.Execute(w, config)
		if err != nil {
			log.Printf("[ ERR ] fill view template error: %v", err)
		}
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
		if repo, err := git.OpenRepository("."); err != nil {
			_, err = git.InitRepository(".", false)
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Printf("git init finished at .")
		} else {
			log.Printf("git repository already found, skip git init")
			repo.Free()
		}
	}
	repo, err := git.OpenRepository(".")
	if err != nil {
		log.Printf("git repository not found at current directory. please use `-init` switch or run `git init` in this directory")
		log.Fatal(err)
		return
	} else {
		repo.Free()
	}
	init_after_main()

	http.HandleFunc("/", handle)

	cnt := 0
	ch := make(chan bool)
	for _, host := range strings.Split(*addr, ",") {
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
