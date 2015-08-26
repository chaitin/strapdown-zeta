package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"strings"
	"time"
)

type CustomOption struct {
	Title         string
	Theme         string
	Toc           string
	HeadingNumber string
	Host          string
}

func (this *RequestContext) SafelyUpdateConfig(path string) {
	option, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(" [ WARN ] " + err.Error())
		return
	}
	var custom_option = CustomOption{}
	err = json.Unmarshal(option, &custom_option)
	if err != nil {
		log.Print(" [ WARN ] " + err.Error())
		return
	}
	if custom_option.Title != "" {
		this.Title = custom_option.Title
		if wikiConfig.verbose {
			log.Print("[ DEBUG ] update page title from config.json")
		}
	}
	if custom_option.HeadingNumber != "" {
		this.HeadingNumber = custom_option.HeadingNumber
		if wikiConfig.verbose {
			log.Print("[ DEBUG ] update page heading number from config.json")
		}
	}
	if custom_option.Toc != "" {
		this.Toc = custom_option.Toc
		if wikiConfig.verbose {
			log.Print("[ DEBUG ] update page toc from config.json")
		}
	}
	if custom_option.Host != "" {
		this.Host = custom_option.Host
		if wikiConfig.verbose {
			log.Print("[ DEBUG ] update page Host from config.json")
		}
	}
	if custom_option.Theme != "" {
		this.Theme = custom_option.Theme
		if wikiConfig.verbose {
			log.Print("[ DEBUG ] update page theme from config.json")
		}
	}
	return
}

func (this *RequestContext) Update() error {
	var comment string
	if this.hasFile {
		comment = "update " + this.path
	} else {
		comment = "upload " + this.path
	}
	// extract the content from post
	upload_content := []byte(this.req.FormValue("body"))
	if len(upload_content) == 0 && this.req.ContentLength > 0 {
		err := this.req.ParseMultipartForm(1048576 * 100)
		if err != nil {
			this.statusCode = http.StatusInternalServerError
			return err
		}
		_, mh, err := this.req.FormFile("body")
		if err != nil {
			this.statusCode = http.StatusBadRequest
			return err
		}
		buffer := &bytes.Buffer{}
		file, err := mh.Open()
		if err != nil {
			this.statusCode = http.StatusBadRequest
			return err
		}
		defer file.Close()
		if _, err = io.Copy(buffer, file); err != nil {
			this.statusCode = http.StatusInternalServerError
			return err
		}
		upload_content = buffer.Bytes()
	}
	// save
	err := SaveAndCommit(this.path, upload_content, comment, "anonymous@"+this.ip)
	if err != nil {
		this.statusCode = http.StatusInternalServerError
		return err
	}
	this.statusCode = http.StatusFound
	http.Redirect(*this.res, this.req, this.req.URL.Path, this.statusCode)
	return nil
}

func (this *RequestContext) View() error {
	custom_view_head, errh := ioutil.ReadFile(this.path + ".head")
	custom_view_tail, errt := ioutil.ReadFile(this.path + ".tail")
	if errh == nil && errt == nil {
		var w = *this.res
		data, err := ioutil.ReadFile(this.path)
		if err != nil {
			w.Write(custom_view_head)
			w.Write(data)
			w.Write(custom_view_tail)
		} else {
			w.Write(custom_view_head)
			w.Write(custom_view_tail)
		}

	} else {
		this.SafelyUpdateConfig(this.path + ".option.json")
		err := templates["view"].Execute(*this.res, this)
		if err != nil {
			log.Printf("[ ERR ] fill view template error: %v", err)
		}
	}
	return nil
}
func (this *RequestContext) Listdir() error {

	w := *this.res
	dirfile, err := SafeOpen(this.path, "")
	if err != nil {
		this.statusCode = http.StatusBadRequest
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	this.SafelyUpdateConfig(this.path + ".option.json")

	if this.Title == wikiConfig.title {
		this.Title = this.path
	}
	this.DirEntries = make([]DirEntry, 0, 16)
	fpstat, err := os.Stat(this.path)
	if err != nil {
		return err
	}
	fpurl := url.URL{Path: path.Join("/", this.path, "..")}
	this.DirEntries = append(this.DirEntries, DirEntry{Name: "..", IsDir: true, Urlpath: fpurl.String(), Size: fpstat.Size(), ModTime: fpstat.ModTime()})

	for {
		dirs, err := dirfile.Readdir(128)
		if err != nil || len(dirs) == 0 {
			break
		}
		for _, d := range dirs {
			dirurl := url.URL{Path: path.Join("/", this.path, d.Name())}
			dirurls := dirurl.String()
			if strings.HasSuffix(dirurls, ".md") {
				dirurls = strings.TrimSuffix(dirurls, ".md")
			}
			this.DirEntries = append(this.DirEntries, DirEntry{Name: d.Name(), IsDir: d.IsDir(), Urlpath: dirurls, Size: d.Size(), ModTime: d.ModTime()})
		}
	}
	return templates["listdir"].Execute(w, this)
}
func (this *RequestContext) History() error {
	commit_history, err := GetHistory(this.path, wikiConfig.histsize)
	if err != nil || commit_history == nil || len(commit_history) == 0 {
		this.statusCode = http.StatusBadRequest
		if err != nil {
			return err
		} else {
			return errors.New("No commit history found for " + this.path)
		}
		return nil
	}
	this.SafelyUpdateConfig(this.path + ".option.json")
	if this.Title == wikiConfig.title {
		this.Title = this.path
	}
	this.CommitEntries = commit_history
	return templates["history"].Execute(*this.res, this)
}
func (this *RequestContext) Edit() error {
	this.SafelyUpdateConfig(this.path + ".option.json")

	return templates["edit"].Execute(*this.res, this)
}
func (this *RequestContext) Diff(diff_ary []string) error {
	// need beatifuy
	var diff string
	var diff_parts []string
	if len(diff_ary) > 0 {
		diff = diff_ary[0]
		diff_parts = strings.Split(diff, ",")
		if len(diff_parts) != 2 {
			return errors.New("Bad Parameter,please select TWO versions!")
		}
	}
	w := *this.res
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	this.SafelyUpdateConfig(this.path + ".option.json")

	content, err := GetFileDiff(this.path, diff_parts)
	this.Content = template.HTML(*content)
	if err != nil {
		this.statusCode = http.StatusBadRequest
		return err
	}
	this.Title = "diff for file from " + diff_parts[0] + " to " + diff_parts[1]
	return templates["diff"].Execute(*this.res, this)
}
func SaveAndCommit(fp string, content []byte, comment string, author string) error {
	var err error

	err = os.MkdirAll(path.Dir(fp), 0700)
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
	defer index.Free()

	err = index.AddByPath(fp)
	if err != nil {
		return err
	}

	treeId, err := index.WriteTree()
	if err != nil {
		return err
	}

	err = index.Write()
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
func GetFileOfVersion(fileName string, version string) ([]byte, error) {
	var err error
	var commit *git.Commit

	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	vl := len(version)

	if vl < 4 || vl > 40 {
		return nil, fmt.Errorf("version length should be in range [4, 40], provided %d", vl)
	}

	oid, err := git.NewOid(version)
	if err == nil {
		// TODO: git2go seems haven't implemented git_commit_lookup_prefix API, so now this lookup only works for full-width 40 hex version
		commit, err = repo.LookupCommit(oid)

		if err == nil && commit != nil {
			str, err := GetFile(repo, commit, fileName)
			if err != nil {
				return nil, err
			}

			var s []byte
			if str != nil {
				s = []byte(*str)
			}
			return s, nil
		}
	}

	// if the commit version id not as long as 40 hexchars, we just loop from head to the initial commit, looking for such a commit matching prefix
	currentBranch, err := repo.Head()
	if err != nil {
		return nil, err
	}
	defer currentBranch.Free()

	commit, err = repo.LookupCommit(currentBranch.Target())
	if err != nil {
		return nil, err
	}

	for commit != nil {
		if commit.Id().String()[0:len(version)] == version {
			str, err := GetFile(repo, commit, fileName)
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
func GetFileDiff(fileName string, diff_versions []string) (*string, error) {
	// only diff .md file
	// diff folder is not supported  or TODO?
	var err error

	// open repo
	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	// get file of diff_versions[0]
	obj0, err := repo.RevparseSingle(fmt.Sprintf("%s:%s", diff_versions[0], fileName))
	if err != nil || obj0 == nil {
		return nil, err
	}
	// get file of diff_versions[1]
	obj1, err := repo.RevparseSingle(fmt.Sprintf("%s:%s", diff_versions[1], fileName))
	if err != nil || obj1 == nil {
		return nil, err
	}
	// TODO: since git2go did not implement
	//           git_diff_blob_to_buffer,git_diff_blobs or git_diff_buffers for sigle file diff
	//           try to use git_diff_tree_to_tree with 2 newly built tree to diff one file
	bld, err := repo.TreeBuilder()
	if err != nil || bld == nil {
		return nil, err
	}
	err = bld.Insert(fileName, obj0.Id(), 0100755)
	if err != nil {
		return nil, err
	}
	treeId1, err := bld.Write()
	if err != nil {
		return nil, err
	}
	// git2go did not implement git_treebuilder_clear,manually remove items
	err = bld.Remove(fileName)
	if err != nil {
		return nil, err
	}
	err = bld.Insert(fileName, obj1.Id(), 0100755)
	if err != nil {
		return nil, err
	}
	treeId2, err := bld.Write()
	if err != nil {
		return nil, err
	}
	defer bld.Free()
	tree1, err := repo.LookupTree(treeId1)
	if err != nil {
		return nil, err
	}
	tree2, err := repo.LookupTree(treeId2)
	if err != nil {
		return nil, err
	}
	// diff,err := repo.DiffTreeToTree(tree1,tree2,nil)
	diff, err := repo.DiffTreeToTree(tree1, tree2, nil)
	if err != nil {
		return nil, err
	}

	diffResult := ""
	filecb := func(diffDelta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		// diffResult += fmt.Sprintf("delta old file: %s new file %s\n",diffDelta.OldFile.Path,diffDelta.NewFile.Path)
		hunkcb := func(diffHunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
			diffResult += fmt.Sprintf("%s", diffHunk.Header)
			linecb := func(diffLine git.DiffLine) error {
				diffPrefix := ""
				switch diffLine.Origin {
				case git.DiffLineAddition:
					diffPrefix = "+"
				case git.DiffLineDeletion:
					diffPrefix = "-"
				}
				diffResult += fmt.Sprintf("%s%s", diffPrefix, diffLine.Content)
				return nil
			}
			return linecb, nil
		}
		return hunkcb, nil
	}

	err = diff.ForEach(filecb, git.DiffDetailLines)
	if err != nil {
		return nil, err
	}

	return &diffResult, nil
}
func GetFile(repo *git.Repository, commit *git.Commit, fileName string) (*string, error) {
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
func GetHistory(fp string, size int) ([]CommitEntry, error) {
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
			if len(filehistory) > 0 && filehistory[len(filehistory)-1].EntryId == entry.Id.String() {
				filehistory = filehistory[:len(filehistory)-1]
			}
			filehistory = append(filehistory, CommitEntry{Id: commit.Id().String(), EntryId: entry.Id.String(), Message: commit.Message(), Author: commit.Author().Name, Timestamp: commit.Author().When})
			cnt += 1
			if size > 0 && len(filehistory) >= size {
				return false
			}
		}
		return true
	})

	return filehistory, nil
}
