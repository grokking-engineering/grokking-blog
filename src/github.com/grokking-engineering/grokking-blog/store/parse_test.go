package store

import (
	"strings"
	"testing"
	"time"
)

var testData = []string{
	`
# Hello world

> 20-10-2016 #foo #bar
>
> Welcome!

Hello!
`, `
# Hello world

    > 20-10-2016
  > Welcome!

Hello!
`, `
# Hello world

    > 20-10-2016

Hello!
`,
}

var expectedArticles = []Article{
	{
		Title:       "Hello world",
		Short:       "Welcome!",
		RawContent:  "Hello!",
		HtmlContent: "<p>Hello!</p>",

		Tags: []string{"foo", "bar"},
		Date: MustParseDate("20-10-2016"),
	},
	{
		Title:       "Hello world",
		Short:       "Welcome!",
		RawContent:  "Hello!",
		HtmlContent: "<p>Hello!</p>",

		Tags: nil,
		Date: MustParseDate("20-10-2016"),
	},
	{
		Title:       "Hello world",
		Short:       "",
		RawContent:  "Hello!",
		HtmlContent: "<p>Hello!</p>",

		Tags: nil,
		Date: MustParseDate("20-10-2016"),
	},
}

var testDataError = [][2]string{
	{`
Hello world
`, "Missing title"},
	{`
# Hello world
`, "Missing date"},
	{`
# Hello world

> Hello
`, "Missing date"},
	{`
# Hello world

> 20-10-2016
`, "Missing content"},
}

func MustParseDate(s string) time.Time {
	t, err := ParseDate(s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestLoadArticle(T *testing.T) {
	for i, input := range testData {
		expected := expectedArticles[i]
		article, err := parseArticle(input)
		if err != nil {
			T.Error("Error parsing article", i, err)
			continue
		}
		if article.Title != expected.Title {
			T.Error("Expect title")
		}
		if article.Date != expected.Date {
			T.Error("Expect date")
		}
		if article.Short != expected.Short {
			T.Error("Expect short")
		}
		if article.RawContent != expected.RawContent {
			T.Error("Expect content")
		}
		if article.HtmlContent != expected.HtmlContent {
			T.Error("Expect html")
		}
		if strings.Join(article.Tags, ",") != strings.Join(expected.Tags, ",") {
			T.Error("Expect tags")
		}
	}
}

func TestLoadArticleError(T *testing.T) {
	for _, testcase := range testDataError {
		input := testcase[0]
		expected := testcase[1]
		_, err := parseArticle(input)
		if err == nil || !strings.Contains(err.Error(), expected) {
			T.Error("Expect error", expected, err)
		}
	}
}
