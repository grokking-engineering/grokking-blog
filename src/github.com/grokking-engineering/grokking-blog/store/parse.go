package store

import (
	"errors"
	"html/template"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

type Article struct {
	Date time.Time
	Tags []string

	Title       string
	Short       string
	RawContent  string
	HtmlContent template.HTML
	Path        template.URL
}

func parseArticle(input string) (*Article, error) {
	p := &parserStruct{}
	article, err := p.parse(input)
	if err != nil {
		return nil, err
	}

	article.HtmlContent = template.HTML(strings.TrimSpace(string(
		blackfriday.MarkdownCommon([]byte(article.RawContent)))))
	return article, nil
}

type parserStruct struct {
	article Article
	input   string

	processingInput string
}

func (p *parserStruct) parse(input string) (*Article, error) {
	parseFuncs := [...]func() error{
		p.parseTitle,
		p.parseInfo,
		p.parseShort,
		p.parseContent,
	}

	p.input = input
	p.processingInput = input
	for step := 0; step < len(parseFuncs); step++ {
		err := parseFuncs[step]()
		if err != nil {
			return nil, err
		}
	}
	return &p.article, nil
}

var (
	ErrEOF      = errors.New("EOF")
	ErrNoPrefix = errors.New("NoPrefix")
	ErrTitle    = errors.New("Missing title")
	ErrDate     = errors.New("Missing date")
	ErrContent  = errors.New("Missing content")
)

func (p *parserStruct) readLine(prefix string) (string, error) {
	for {
		index := strings.Index(p.processingInput, "\n")
		if index < 0 {
			return "", ErrEOF
		}
		line := p.processingInput[:index]
		line = strings.TrimSpace(line)
		if line == "" {
			p.processingInput = p.processingInput[index+1:]
			continue
		}

		if strings.Index(line, prefix) != 0 {
			// Do not process the input here!
			return "", ErrNoPrefix
		}
		line = strings.TrimSpace(line[len(prefix):])

		p.processingInput = p.processingInput[index+1:]
		if line == "" {
			continue
		}
		return line, nil
	}
}

func (p *parserStruct) parseTitle() error {
	line, err := p.readLine("#")
	if err != nil {
		return ErrTitle
	}
	p.article.Title = line
	return nil
}

func (p *parserStruct) parseInfo() error {
	line, err := p.readLine(">")
	if err != nil {
		return ErrDate
	}

	parts := strings.Split(line, " ")
	if len(parts) == 0 {
		return ErrDate
	}

	date, err := ParseDate(parts[0])
	if err != nil {
		return ErrDate
	}

	parts = parts[1:]
	tags := make([]string, len(parts))[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.Index(p, "#") == 0 {
			tags = append(tags, p[1:])
		}
	}

	p.article.Date = date
	if len(tags) > 0 {
		p.article.Tags = tags
	}
	return nil
}

func (p *parserStruct) parseShort() error {
	short := ""
	for {
		line, err := p.readLine(">")
		if err != nil {
			break
		}
		if short != "" {
			short += "\n"
		}
		short += line
	}
	p.article.Short = short
	return nil
}

func (p *parserStruct) parseContent() error {
	content := strings.TrimSpace(p.processingInput)
	if content == "" {
		return ErrContent
	}
	p.article.RawContent = content
	return nil
}
