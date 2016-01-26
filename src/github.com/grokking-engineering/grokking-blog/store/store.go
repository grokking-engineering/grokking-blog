package store

import (
	"html/template"
	"time"

	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

var kTimeFormat = "02-01-2006"

var l = logs.New("store")

func ParseDate(s string) (time.Time, error) {
	return time.Parse(kTimeFormat, s)
}

type Instance struct {
	ContentDir string

	data *Data
}

func (this *Instance) Init() {
	if this.ContentDir == "" {
		panic("Empty ContentDir")
	}
	err := this.ClearCacheAndReload()
	if err != nil {
		l.Fatal(err)
	}
}

func (this *Instance) ClearCacheAndReload() error {
	data, err := loadFiles(this.ContentDir)
	if err != nil {
		l.WithError(err).Error("Unable to load content!")
		return err
	}

	this.data = data
	l.Println("Loaded content")
	for k := range data.Entries {
		l.Println("Indexed:", k)
	}
	return nil
}

func (this *Instance) GetEntry(path string) *Entry {
	entry := this.data.Entries[path]
	return entry
}

func (this *Instance) GetDir(path string) *Dir {
	dir := this.data.Dirs[path]
	return dir
}

func (this *Instance) GetMainLayout() *template.Template {
	return this.data.MainLayout
}
