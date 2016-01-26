# Grokking Blog

## Quick Start

Make sure you have [Go 1.5.3](https://golang.org/doc/install) installed and `go` in your `$PATH`.

```
git clone https://github.com/grokking-engineering/grokking-blog
cd grokking-blog
export GOPATH=$PWD
git submodule init
git submodule update
go install github.com/grokking-engineering/grokking-blog
bin/grokking-blog
```

Then open http://localhost:8080

### Production

```
export DEVELOPMENT=0
bin/grokking-blog
```

## Syntax

### Directory tree

```
content/
  _layout.tpl.html        // (required) article layout,     
  _layout_main.tpl.html   // (required) top level layout    
  index.md                // (required) top level article 
  index.tpl.html          // (optional) layout for index.md
                          // fallback to _layout.tpl.html
  <article>.md            // access at: /article
  <article>.tpl.html      // (optional) fallback to _layout.tpl.html

  <dirname>/
    _layout.tpl.html      // (optional) fallback to upper _layout.tpl.html
    index.md              // access at: /dirname/
    index.tpl.html        // (optional)
    <article>.md          // access at: /dirname/article
    <article>.tpl.html    // (optional)
```

### Markdown syntax

```
# Title

> date #tag1 #tag2
>
> Short description

Markdown Content
```

### Template syntax

**article.md**

```
  {{.Title}}         // Title
  {{.HtmlContent}}   // Content
  {{.Path}}          // Relative url
  {{.Date}}          // Date
```

**index.md**

```
  // List all articles in directory "blog"
  // ("." for current directory)
  {{range dir "blog"}}

  {{else}}

  {{end}}
```
