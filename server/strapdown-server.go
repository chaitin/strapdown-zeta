package main

import (
	"flag"
	"fmt"
	"github.com/libgit2/git2go"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var port = flag.Int("port", 8080, "The port for the server to listen")
var addr = flag.String("address", "0.0.0.0", "Listening address")
var view_head, view_tail, edit_head, edit_tail []byte

func init() {
	var err error
	if view_head, err = ioutil.ReadFile("view.head"); err != nil {
		log.Fatalf("cannot read view.head")
	}
	if view_tail, err = ioutil.ReadFile("view.tail"); err != nil {
		log.Fatalf("cannot read view.tail")
	}
	if edit_head, err = ioutil.ReadFile("edit.head"); err != nil {
		log.Fatalf("cannot read edit.head")
	}
	if edit_tail, err = ioutil.ReadFile("edit.tail"); err != nil {
		log.Fatalf("cannot read edit.tail")
	}
	edit_head = []byte(strings.TrimSpace(string(edit_head)))
	edit_tail = []byte(strings.TrimSpace(string(edit_tail)))
}

func push(fp string, content []byte, comment string, author string) error {
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

func handle(w http.ResponseWriter, r *http.Request) {
	fp := r.URL.Path[1:] + ".md"

	if r.Method == "POST" || r.Method == "PUT" {
		err := push(fp, []byte(r.FormValue("body")), "update "+fp, "anonymous@"+remote_ip(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
		return
	}

	content, err := ioutil.ReadFile(fp)

	handleEdit := func() {
		w.Write(edit_head)
		w.Write(content)
		w.Write(edit_tail)
	}

	if err != nil {
		if _, err := os.Stat(fp); err != nil {
			// file not exist or permission denied, enter edit mode
			handleEdit()
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	q := r.URL.Query()
	_, exists := q["edit"]

	if exists {
		// enter edit mode
		handleEdit()
		return
	}

	custom_view_head, err := ioutil.ReadFile(fp + ".head")
	if err != nil {
		w.Write(view_head)
	} else {
		w.Write(custom_view_head)
	}
	w.Write(content)
	custom_view_tail, err := ioutil.ReadFile(fp + ".tail")
	if err != nil {
		w.Write(view_tail)
	} else {
		w.Write(custom_view_tail)
	}
}

func main() {
	flag.Parse()
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
