package converter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"

	"github.com/ralvarez/m2p/internal/assets"
)

type templateData struct {
	Title          string
	Body           template.HTML
	Date           string
	ShowFooter     bool
	PageBreakLevel int
	CSS            template.CSS
}

var (
	docTmpl = mustLoadTemplate()
	baseCSS = mustLoadCSS()
	fontCSS = buildFontCSS()
)

func mustLoadTemplate() *template.Template {
	t, err := template.New("document").Parse(assets.DocumentHTML)
	if err != nil {
		panic(fmt.Sprintf("assets: invalid document.html template: %v", err))
	}
	return t
}

func mustLoadCSS() template.CSS {
	return template.CSS(assets.StylesCSS)
}

// buildFontCSS encodes embedded WOFF2 files and returns @font-face CSS with
// base64 data URIs so the rendered HTML is fully self-contained.
func buildFontCSS() template.CSS {
	interB64 := base64.StdEncoding.EncodeToString(assets.InterFont)
	jbmB64 := base64.StdEncoding.EncodeToString(assets.JetBrainsMonoFont)
	return template.CSS(fmt.Sprintf(interFontTemplate+"\n"+jetBrainsMonoTemplate, interB64, jbmB64))
}

// RenderTemplate injects the HTML fragment into the full document template.
func RenderTemplate(fragment []byte, title, date string, showFooter bool, pageBreakLevel int) ([]byte, error) {
	data := templateData{
		Title:          title,
		Body:           template.HTML(fragment), //nolint:gosec // fragment comes from goldmark, not user input
		Date:           date,
		ShowFooter:     showFooter,
		PageBreakLevel: pageBreakLevel,
		CSS:            fontCSS + baseCSS,
	}
	var buf bytes.Buffer
	if err := docTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
