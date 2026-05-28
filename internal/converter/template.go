package converter

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/ralvarez/m2p/internal/assets"
)

type templateData struct {
	Title      string
	Body       template.HTML
	Date       string
	ShowFooter bool
}

var docTmpl = mustLoadTemplate()

func mustLoadTemplate() *template.Template {
	raw, err := assets.FS.ReadFile("templates/document.html")
	if err != nil {
		panic(fmt.Sprintf("assets: missing document.html: %v", err))
	}
	t, err := template.New("document").Parse(string(raw))
	if err != nil {
		panic(fmt.Sprintf("assets: invalid document.html template: %v", err))
	}
	return t
}

// RenderTemplate injects the HTML fragment into the full document template.
func RenderTemplate(fragment []byte, title, date string, showFooter bool) ([]byte, error) {
	data := templateData{
		Title:      title,
		Body:       template.HTML(fragment), //nolint:gosec // fragment comes from goldmark, not user input
		Date:       date,
		ShowFooter: showFooter,
	}
	var buf bytes.Buffer
	if err := docTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
