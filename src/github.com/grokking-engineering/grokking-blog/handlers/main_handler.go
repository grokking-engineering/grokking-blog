package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/grokking-engineering/grokking-blog/store"
	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

var l = logs.New("handlers")

type MainHandler struct {
	Store *store.Instance
	IsDev bool
}

func (this *MainHandler) Init() {
	if this.Store == nil {
		panic("Required object is nil")
	}
}

func (this *MainHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if this.IsDev {
		err := this.Store.ClearCacheAndReload()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "DEVELOPMENT MODE: Unable to reload content. Please check server log:", err)
			return
		}
	}

	entryPath, err := filepath.Rel("/", req.URL.Path)
	if err != nil {
		this.notFound(w)
		return
	}

	l.WithFields(logs.M{
		"entryPath": entryPath,
	}).Info("Serve entry")
	entry := this.Store.GetEntry(entryPath)
	if entry == nil {
		this.notFound(w)
		return
	}

	if entry.IsDir && !strings.HasSuffix(req.URL.Path, "/") {
		http.Redirect(w, req, req.URL.Path+"/", http.StatusMovedPermanently)
		return
	}

	this.renderEntry(w, entry)
}

func (this *MainHandler) notFound(w http.ResponseWriter) {
	mainLayout := this.Store.GetMainLayout()
	w.WriteHeader(http.StatusNotFound)
	must(mainLayout.Execute(w, "404 Not Found"))
}

func (this *MainHandler) serverError(w http.ResponseWriter) {
	mainLayout := this.Store.GetMainLayout()
	w.WriteHeader(http.StatusInternalServerError)
	must(mainLayout.Execute(w, "500 Server Error"))
}

func (this *MainHandler) renderEntry(w http.ResponseWriter, entry *store.Entry) {
	buf := &bytes.Buffer{}

	err := entry.Layout.Execute(buf, entry.Article)
	if err != nil {
		l.WithError(err).Error("renderEntry")
		this.serverError(w)
		return
	}

	mainLayout := this.Store.GetMainLayout()
	must(mainLayout.Execute(w, template.HTML(buf.String())))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
