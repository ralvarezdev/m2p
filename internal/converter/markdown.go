package converter

import (
	"bytes"
	"regexp"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Table,
		highlighting.NewHighlighting(
			highlighting.WithStyle(highlightTheme),
			highlighting.WithFormatOptions(
				chromahtml.WithLineNumbers(false),
			),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
)

var h1RE = regexp.MustCompile(`(?m)^#\s+(.+)$`)

// ParseMarkdown converts Markdown source to an HTML fragment and extracts the
// document title from the first h1 heading (if present).
func ParseMarkdown(src []byte) (fragment []byte, title string, err error) {
	if m := h1RE.Find(src); m != nil {
		title = string(h1RE.FindSubmatch(src)[1])
	}

	var buf bytes.Buffer
	if err = md.Convert(src, &buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), title, nil
}
