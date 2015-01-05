package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
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
}

func handle(w http.ResponseWriter, r *http.Request) {
	fp := r.URL.Path[1:] + ".md"

	if r.Method == "POST" || r.Method == "PUT" {
		err := ioutil.WriteFile(fp, []byte(r.FormValue("body")), 0600)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
		return
	}

	content, err := ioutil.ReadFile(fp)

	handleEdit := func() {
		fmt.Fprintf(w, "%s\n", edit_head)
		w.Write(content)
		fmt.Fprintf(w, "\n%s", edit_tail)
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

	fmt.Fprintf(w, "%s\n", view_head)
	w.Write(content)
	fmt.Fprintf(w, "\n%s", view_tail)
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
