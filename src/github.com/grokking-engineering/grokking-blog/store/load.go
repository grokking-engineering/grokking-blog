package store

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

type Data struct {
	MainLayout *template.Template

	Entries map[string]*Entry
	Dirs    map[string]*Dir

	SortedArticles []*Article
}

type Entry struct {
	Article *Article
	Layout  *template.Template
	IsDir   bool
}

type Dir struct {
	Path    string
	Layout  *template.Template
	Entries map[string]*Entry

	SortedArticles []*Article
}

func loadFiles(rootDir string) (*Data, error) {

	data := &Data{
		Entries: make(map[string]*Entry),
		Dirs:    make(map[string]*Dir),
	}

	makeFuncMap := func(basePath string) template.FuncMap {
		return template.FuncMap{
			"dir": func(dirPath string) ([]*Article, error) {
				relDirPath, err := filepath.Rel(rootDir, filepath.Join(basePath, dirPath))
				if err != nil {
					return nil, err
				}
				dir := data.Dirs[relDirPath]
				if dir == nil {
					return nil, errors.New("DirPath not exist: " + relDirPath)
				}
				return dir.SortedArticles, nil
			},
		}
	}

	parseFiles := func(path string) (*template.Template, error) {
		tpl := template.New(filepath.Base(path))
		tpl.Funcs(makeFuncMap(filepath.Dir(path)))
		return tpl.ParseFiles(path)
	}

	// load main layout
	mainLayoutPath := filepath.Join(rootDir, "_layout_main.tpl.html")
	tpl, err := parseFiles(mainLayoutPath)
	if err != nil {
		l.WithError(err).WithFields(logs.M{
			"mainLayoutPath": mainLayoutPath,
		}).Error("Unable to load main layout")
		return nil, errors.New("Fatal")
	}
	data.MainLayout = tpl

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// load dir
		if info.IsDir() {
			log.Println("Load dir: ", relativePath)
			dir := &Dir{Path: relativePath}
			dir.Entries = make(map[string]*Entry)
			data.Dirs[relativePath] = dir

			layoutPath := filepath.Join(path, "_layout.tpl.html")
			_, err := os.Stat(layoutPath)
			if err != nil {
				// Skip reading template file
				return nil
			}

			// load dir template
			log.Println("Load tpl: ", layoutPath)
			tpl, err := parseFiles(layoutPath)
			if err != nil {
				l.WithError(err).WithFields(logs.M{
					"layoutPath": layoutPath,
				}).Error("Unable to parse template!")
				return errors.New("Fatal")
			}

			dir.Layout = tpl
			return nil
		}

		ext := filepath.Ext(relativePath)
		if ext != ".md" {
			return nil
		}
		baseName := filepath.Base(relativePath)
		baseName = baseName[:len(baseName)-len(ext)]
		stripPath := relativePath[:len(relativePath)-len(ext)]
		dirPath := filepath.Dir(relativePath)

		log.Println("Load file:", relativePath)
		entry := &Entry{}

		// save to entries, check for index.md
		if baseName == "index" {
			data.Entries[dirPath] = entry
			entry.IsDir = true
		} else {
			data.Entries[stripPath] = entry
		}

		// load article
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			l.WithError(err).WithFields(logs.M{
				"path": path,
			}).Error("Unable to load .md file!")
			return errors.New("Fatal")
		}

		article, err := parseArticle(string(bytes))
		if err != nil {
			l.WithError(err).WithFields(logs.M{
				"path": path,
			}).Error("Unable to parse .md file!")
			return errors.New("Fatal")
		}

		article.Path = template.URL(stripPath)
		entry.Article = article

		// load article template

		layoutBaseName := baseName + ".tpl.html"
		layoutPath := filepath.Join(filepath.Dir(path), layoutBaseName)
		_, err = os.Stat(layoutPath)
		if err != nil {
			log.Println("Skip tpl: ", layoutPath)
		} else {
			log.Println("Load tpl: ", layoutPath)
			tpl, err := parseFiles(layoutPath)
			if err != nil {
				l.WithError(err).WithFields(logs.M{
					"path": layoutPath,
				}).Error("Unable to parse template!")
				return errors.New("Fatal")
			}

			entry.Layout = tpl
		}

		// update dir info
		dir := data.Dirs[dirPath]
		if dir != nil {
			if !entry.IsDir {
				dir.Entries[relativePath] = entry
			}
		} else {
			l.WithFields(logs.M{
				"dirPath":      dirPath,
				"relativePath": relativePath,
			}).Error("Something is wrong: dirPath not exist!")
			return errors.New("Fatal")
		}

		return nil
	}

	err = filepath.Walk(rootDir, walkFunc)
	if err != nil {
		return nil, err
	}

	// clean up
	for path, entry := range data.Entries {
		if entry.Layout == nil {
			entry.Layout = inheritLayout(data, path)
			if entry.Layout == nil {
				l.WithFields(logs.M{
					"path": path,
				}).Error("No template found for article")
				return nil, errors.New("Fatal")
			}
		}
	}

	data.SortedArticles = getSortedArticles(data.Entries)
	for _, dir := range data.Dirs {
		dir.SortedArticles = getSortedArticles(dir.Entries)
	}

	return data, nil
}

func inheritLayout(data *Data, path string) *template.Template {
	for {
		dirPath := filepath.Dir(path)
		dir := data.Dirs[dirPath]
		if dir == nil {
			return nil
		}
		if dir.Layout != nil {
			return dir.Layout
		}
		if dirPath == "." {
			return nil
		}
	}
}

type articleByDate []*Article

func (a articleByDate) Len() int      { return len(a) }
func (a articleByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a articleByDate) Less(i, j int) bool {
	return a[i].Date.Sub(a[j].Date) < 0
}

func getSortedArticles(entries map[string]*Entry) []*Article {
	var a []*Article
	for _, entry := range entries {
		a = append(a, entry.Article)
	}
	sort.Sort(articleByDate(a))
	return a
}
