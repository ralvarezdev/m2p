package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// nativeRenderer produces PDFs using a pure-Go fpdf backend.
// It walks the goldmark AST and maps each node to fpdf drawing calls,
// approximating the CSS design without requiring any external browser.
type nativeRenderer struct{}

func (r *nativeRenderer) Render(_ context.Context, in RenderInput) error {
	p := nativePaperSize(in.Paper)
	pdf := fpdf.New("P", "mm", p, "")
	margin := float64(NativeMarginDefault)
	pdf.SetMargins(margin, margin, margin)

	bottomBreakMargin := float64(NativeMarginDefault)
	if in.ShowFooter {
		bottomBreakMargin = float64(NativeMarginFooter)
	}
	pdf.SetAutoPageBreak(true, bottomBreakMargin)

	w := &nativeWriter{pdf: pdf, src: in.Source, pageBreakLevel: in.PageBreakLevel}

	if in.ShowFooter {
		date := in.Date
		pdf.SetFooterFunc(func() { w.drawFooter(date) })
	}

	pdf.AddPage()
	w.initFonts()

	if in.Title != "" {
		w.drawTitle(in.Title)
	}

	mdParser := nativeMDParser()
	reader := text.NewReader(in.Source)
	doc := mdParser.Parse(reader)

	if err := ast.Walk(doc, w.walk); err != nil {
		return fmt.Errorf("ast walk: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(in.Output), DirPerm); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return pdf.OutputFileAndClose(in.Output)
}

func nativePaperSize(p Paper) string {
	switch p {
	case PaperLetter:
		return "Letter"
	case PaperA3:
		return "A3"
	case PaperLegal:
		return "Legal"
	default:
		return "A4"
	}
}

// nativeMDParser returns a goldmark parser with GFM table support (AST only, no HTML output).
func nativeMDParser() parser.Parser {
	return goldmark.New(
		goldmark.WithExtensions(extension.Table),
	).Parser()
}

type nativeWriter struct {
	pdf             *fpdf.Fpdf
	src             []byte
	listOrdered     []bool
	listCounters    []int
	pageBreakLevel  int
	breakLevelCount int
	listDepth       int
}

// inlineRun holds a segment of inline text with its formatting state.
type inlineRun struct {
	text   string
	bold   bool
	italic bool
	mono   bool
}

// Design tokens matched to the CSS concept.
type rgbColor struct{ r, g, b int }

var (
	nColFG       = rgbColor{26, 28, 46}
	nColMuted    = rgbColor{74, 80, 120}
	nColFaint    = rgbColor{142, 148, 181}
	nColBorder   = rgbColor{200, 202, 214}
	nColH1Line   = rgbColor{122, 162, 247}
	nColAccent2  = rgbColor{90, 74, 140}
	nColCodeBG   = rgbColor{26, 28, 46}
	nColCodeDot  = rgbColor{58, 63, 98}
	nColQuoteBG  = rgbColor{245, 242, 252}
	nColQuoteBdr = rgbColor{187, 154, 247}
)

func (w *nativeWriter) setColor(c rgbColor) { w.pdf.SetTextColor(c.r, c.g, c.b) }

func (w *nativeWriter) initFonts() {
	w.pdf.SetFont("Helvetica", "", float64(TypoBody))
	w.setColor(nColFG)
}

func (w *nativeWriter) walk(n ast.Node, entering bool) (ast.WalkStatus, error) {
	switch node := n.(type) {
	case *ast.Heading:
		if entering {
			w.renderHeading(node)
			return ast.WalkSkipChildren, nil
		}
	case *ast.Paragraph:
		if entering {
			w.renderParagraph(node)
			return ast.WalkSkipChildren, nil
		}
	case *ast.FencedCodeBlock:
		if entering {
			w.renderCodeBlock(node)
			return ast.WalkSkipChildren, nil
		}
	case *ast.CodeBlock:
		if entering {
			w.renderCodeBlock(node)
			return ast.WalkSkipChildren, nil
		}
	case *ast.Blockquote:
		if entering {
			w.renderBlockquote(node)
			return ast.WalkSkipChildren, nil
		}
	case *ast.List:
		if entering {
			w.listDepth++
			w.listOrdered = append(w.listOrdered, node.IsOrdered())
			w.listCounters = append(w.listCounters, 0)
		} else {
			w.listDepth--
			w.listOrdered = w.listOrdered[:len(w.listOrdered)-1]
			w.listCounters = w.listCounters[:len(w.listCounters)-1]
			if w.listDepth == 0 {
				w.pdf.Ln(3)
			}
		}
	case *ast.ListItem:
		if entering {
			w.renderListItem(node)
			return ast.WalkSkipChildren, nil
		}
	case *east.Table:
		if entering {
			w.renderTable(node)
			return ast.WalkSkipChildren, nil
		}
	case *ast.ThematicBreak:
		if entering {
			w.renderHR()
		}
	}
	return ast.WalkContinue, nil
}

func (w *nativeWriter) drawTitle(title string) {
	w.pdf.SetFont("Helvetica", "B", float64(TypoH1))
	w.setColor(nColFG)
	w.pdf.MultiCell(0, 10, title, "", "L", false)

	y := w.pdf.GetY()
	pw, _ := w.pdf.GetPageSize()
	lm, _, rm, _ := w.pdf.GetMargins()
	w.pdf.SetDrawColor(nColH1Line.r, nColH1Line.g, nColH1Line.b)
	w.pdf.SetLineWidth(0.7)
	w.pdf.Line(lm, y, pw-rm, y)
	w.pdf.SetDrawColor(0, 0, 0)
	w.pdf.SetLineWidth(0.2)
	w.pdf.Ln(6)
}

func (w *nativeWriter) renderHeading(node *ast.Heading) {
	if w.pageBreakLevel > 0 && node.Level == w.pageBreakLevel {
		if w.breakLevelCount > 0 {
			w.pdf.AddPage()
		}
		w.breakLevelCount++
	}

	content := w.collectText(node)
	pw, _ := w.pdf.GetPageSize()
	lm, _, rm, _ := w.pdf.GetMargins()

	switch node.Level {
	case 1:
		w.pdf.Ln(6)
		w.pdf.SetFont("Helvetica", "B", float64(TypoH2))
		w.setColor(nColFG)
		w.pdf.MultiCell(0, 9, content, "", "L", false)
		y := w.pdf.GetY()
		w.pdf.SetDrawColor(nColH1Line.r, nColH1Line.g, nColH1Line.b)
		w.pdf.SetLineWidth(0.6)
		w.pdf.Line(lm, y, pw-rm, y)
		w.pdf.SetDrawColor(0, 0, 0)
		w.pdf.SetLineWidth(0.2)
		w.pdf.Ln(5)

	case 2:
		w.pdf.Ln(8)
		w.pdf.SetFont("Courier", "", 9)
		w.setColor(nColAccent2)
		w.pdf.CellFormat(5, 7, "§", "", 0, "L", false, 0, "")
		w.pdf.SetFont("Helvetica", "B", float64(TypoH3))
		w.setColor(nColFG)
		w.pdf.MultiCell(0, 7, content, "", "L", false)
		w.pdf.Ln(2)

	case 3:
		w.pdf.Ln(6)
		w.pdf.SetFont("Courier", "B", float64(TypoCode))
		w.setColor(nColMuted)
		w.pdf.MultiCell(0, 6, strings.ToUpper(content), "", "L", false)
		w.pdf.Ln(2)

	default:
		w.pdf.Ln(4)
		w.pdf.SetFont("Helvetica", "B", float64(TypoBody))
		w.setColor(nColFG)
		w.pdf.MultiCell(0, 6, content, "", "L", false)
		w.pdf.Ln(1)
	}

	w.pdf.SetFont("Helvetica", "", float64(TypoBody))
	w.setColor(nColFG)
}

func (w *nativeWriter) renderParagraph(node *ast.Paragraph) {
	lm, _, rm, _ := w.pdf.GetMargins()
	pw, _ := w.pdf.GetPageSize()
	lineW := pw - lm - rm

	runs := w.buildRuns(node, false, false, false)
	textContent := runsToString(runs)

	w.pdf.SetFont("Helvetica", "", float64(TypoBody))
	w.setColor(nColFG)
	w.pdf.MultiCell(lineW, 5.5, textContent, "", "L", false)
	w.pdf.Ln(3)
}

func (w *nativeWriter) renderCodeBlock(node ast.Node) {
	// Detect language from fenced code block info string.
	lang := ""
	if fb, ok := node.(*ast.FencedCodeBlock); ok {
		if lb := fb.Language(w.src); lb != nil {
			lang = strings.TrimSpace(string(lb))
		}
	}

	// Collect raw source lines.
	var rawLines []string
	for i := 0; i < node.Lines().Len(); i++ {
		seg := node.Lines().At(i)
		rawLines = append(rawLines, strings.TrimRight(string(seg.Value(w.src)), "\n"))
	}
	content := strings.Join(rawLines, "\n")

	// Tokenize with chroma.
	tokens, _ := tokenize(lang, content)
	tokenLines := splitTokensByLine(tokens)
	if len(tokenLines) == 0 {
		tokenLines = [][]chroma.Token{{}}
	}

	w.pdf.Ln(3)
	lm, _, rm, _ := w.pdf.GetMargins()
	pw, _ := w.pdf.GetPageSize()
	boxW := pw - lm - rm

	const lineH = 4.5
	boxH := float64(len(tokenLines))*lineH + float64(CodeBlockPaddingY)

	startX, startY := w.pdf.GetX(), w.pdf.GetY()

	// Background
	w.pdf.SetFillColor(nColCodeBG.r, nColCodeBG.g, nColCodeBG.b)
	w.pdf.Rect(startX, startY, boxW, boxH, "F")

	// Decorative ● ● ●
	w.pdf.SetFont("Courier", "", 6)
	w.pdf.SetTextColor(nColCodeDot.r, nColCodeDot.g, nColCodeDot.b)
	w.pdf.SetXY(startX+boxW-float64(CodeBlockMarginR), startY+float64(CodeBlockMarginT))
	w.pdf.CellFormat(float64(CodeBlockHeaderW), float64(CodeBlockHeaderH), "● ● ●", "", 0, "R", false, 0, "")

	// Render token by token, line by line.
	w.pdf.SetFont("Courier", "", 9)
	for li, line := range tokenLines {
		curX := startX + 5
		curY := startY + 5 + float64(li)*lineH
		for _, tok := range line {
			r, g, b := tokenRGB(tok.Type)
			w.pdf.SetTextColor(int(r), int(g), int(b))
			w.pdf.SetXY(curX, curY)
			tokW := w.pdf.GetStringWidth(tok.Value)
			w.pdf.CellFormat(tokW, lineH, tok.Value, "", 0, "L", false, 0, "")
			curX += tokW
		}
	}

	w.pdf.SetXY(lm, startY+boxH)
	w.pdf.Ln(5)
	w.pdf.SetFont("Helvetica", "", float64(TypoBody))
	w.setColor(nColFG)
}

func (w *nativeWriter) renderBlockquote(node *ast.Blockquote) {
	var texts []string
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if p, ok := child.(*ast.Paragraph); ok {
			texts = append(texts, w.collectText(p))
		}
	}
	content := strings.Join(texts, "\n")

	w.pdf.Ln(3)
	lm, _, rm, _ := w.pdf.GetMargins()
	pw, _ := w.pdf.GetPageSize()
	boxW := pw - lm - rm

	w.pdf.SetFont("Helvetica", "I", float64(TypoCode))
	splitLines := w.pdf.SplitLines([]byte(content), boxW-float64(CodeBlockPaddingX))
	boxH := float64(len(splitLines))*5.5 + float64(CodeBlockPaddingY)

	x, y := w.pdf.GetX(), w.pdf.GetY()
	w.pdf.SetFillColor(nColQuoteBG.r, nColQuoteBG.g, nColQuoteBG.b)
	w.pdf.Rect(x, y, boxW, boxH, "F")

	w.pdf.SetDrawColor(nColQuoteBdr.r, nColQuoteBdr.g, nColQuoteBdr.b)
	w.pdf.SetLineWidth(1.2)
	w.pdf.Line(x, y, x, y+boxH)
	w.pdf.SetLineWidth(0.2)
	w.pdf.SetDrawColor(0, 0, 0)

	w.setColor(nColMuted)
	w.pdf.SetXY(x+7, y+5)
	w.pdf.MultiCell(boxW-12, 5.5, content, "", "L", false)

	w.pdf.SetXY(lm, y+boxH)
	w.pdf.Ln(4)
	w.pdf.SetFont("Helvetica", "", float64(TypoBody))
	w.setColor(nColFG)
}

func (w *nativeWriter) renderListItem(node *ast.ListItem) {
	depth := w.listDepth - 1
	indent := float64(depth) * 6

	ordered := len(w.listOrdered) > 0 && w.listOrdered[len(w.listOrdered)-1]
	var bullet string
	if ordered {
		w.listCounters[len(w.listCounters)-1]++
		bullet = fmt.Sprintf("%d.", w.listCounters[len(w.listCounters)-1])
	} else {
		bullet = "-"
	}

	lm, _, rm, _ := w.pdf.GetMargins()
	pw, _ := w.pdf.GetPageSize()
	bulletW := 6.0

	w.pdf.SetFont("Helvetica", "", 11)
	w.setColor(nColFaint)
	w.pdf.SetX(lm + indent)
	w.pdf.CellFormat(bulletW, 6, bullet, "", 0, "R", false, 0, "")

	var lines []string
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if p, ok := child.(*ast.Paragraph); ok {
			lines = append(lines, w.collectText(p))
		}
	}
	itemText := strings.Join(lines, " ")

	w.setColor(nColFG)
	textW := pw - lm - rm - indent - bulletW - 2
	w.pdf.SetX(lm + indent + bulletW + 2)
	w.pdf.MultiCell(textW, 6, itemText, "", "L", false)
}

func (w *nativeWriter) renderTable(node *east.Table) {
	w.pdf.Ln(3)
	lm, _, rm, _ := w.pdf.GetMargins()
	pw, _ := w.pdf.GetPageSize()
	tableW := pw - lm - rm

	type tableRow struct {
		cells    []string
		isHeader bool
	}
	var rows []tableRow

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch rowNode := child.(type) {
		case *east.TableHeader:
			r := tableRow{isHeader: true}
			for cell := rowNode.FirstChild(); cell != nil; cell = cell.NextSibling() {
				r.cells = append(r.cells, w.collectText(cell))
			}
			rows = append(rows, r)
		case *east.TableRow:
			r := tableRow{}
			for cell := rowNode.FirstChild(); cell != nil; cell = cell.NextSibling() {
				r.cells = append(r.cells, w.collectText(cell))
			}
			rows = append(rows, r)
		}
	}

	if len(rows) == 0 {
		return
	}
	cols := 0
	for _, r := range rows {
		if len(r.cells) > cols {
			cols = len(r.cells)
		}
	}
	if cols == 0 {
		return
	}
	colW := tableW / float64(cols)

	for _, r := range rows {
		if r.isHeader {
			w.pdf.SetFont("Courier", "B", 8)
			w.setColor(nColFaint)
			w.pdf.SetDrawColor(nColBorder.r, nColBorder.g, nColBorder.b)
			for _, cell := range r.cells {
				w.pdf.CellFormat(colW, 7, strings.ToUpper(cell), "B", 0, "L", false, 0, "")
			}
		} else {
			w.pdf.SetFont("Helvetica", "", 10)
			w.setColor(nColFG)
			w.pdf.SetDrawColor(nColBorder.r, nColBorder.g, nColBorder.b)
			for _, cell := range r.cells {
				w.pdf.CellFormat(colW, 7, cell, "B", 0, "L", false, 0, "")
			}
		}
		w.pdf.Ln(-1)
	}

	w.pdf.SetDrawColor(0, 0, 0)
	w.pdf.Ln(4)
	w.pdf.SetFont("Helvetica", "", 11)
	w.setColor(nColFG)
}

func (w *nativeWriter) renderHR() {
	w.pdf.Ln(4)
	lm, _, rm, _ := w.pdf.GetMargins()
	pw, _ := w.pdf.GetPageSize()
	y := w.pdf.GetY()
	w.pdf.SetDrawColor(nColBorder.r, nColBorder.g, nColBorder.b)
	w.pdf.Line(lm, y, pw-rm, y)
	w.pdf.SetDrawColor(0, 0, 0)
	w.pdf.Ln(6)
}

func (w *nativeWriter) drawFooter(date string) {
	pw, ph := w.pdf.GetPageSize()
	_, _, rm, bm := w.pdf.GetMargins()
	y := ph - bm - 5

	w.pdf.SetDrawColor(nColBorder.r, nColBorder.g, nColBorder.b)
	lm := 20.0
	w.pdf.Line(lm, y-2, pw-rm, y-2)
	w.pdf.SetDrawColor(0, 0, 0)

	w.pdf.SetFont("Courier", "", 8)
	w.setColor(nColFaint)
	w.pdf.SetXY(lm, y)
	w.pdf.CellFormat(0, 5, "0xRA · m2p", "", 0, "L", false, 0, "")
	w.pdf.SetXY(lm, y)
	w.pdf.CellFormat(pw-lm-rm, 5, date, "", 0, "R", false, 0, "")

	// Restore body font so content on the next page starts correctly.
	w.pdf.SetFont("Helvetica", "", 11)
	w.setColor(nColFG)
}

// collectText flattens all inline text under a node into a plain string.
func (w *nativeWriter) collectText(n ast.Node) string {
	var sb strings.Builder
	collectTextInto(n, w.src, &sb)
	return strings.TrimSpace(sb.String())
}

func collectTextInto(n ast.Node, src []byte, sb *strings.Builder) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			sb.Write(t.Segment.Value(src))
			if t.SoftLineBreak() {
				sb.WriteByte(' ')
			}
		} else {
			collectTextInto(child, src, sb)
		}
	}
	// Handle leaf text nodes without children
	if t, ok := n.(*ast.Text); ok {
		sb.Write(t.Segment.Value(src))
	}
}

// buildRuns collects inline content as a slice of styled runs.
func (w *nativeWriter) buildRuns(n ast.Node, bold, italic, mono bool) []inlineRun {
	var runs []inlineRun
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *ast.Text:
			t := string(node.Segment.Value(w.src))
			if node.SoftLineBreak() {
				t += " "
			}
			if utf8.ValidString(t) && t != "" {
				runs = append(runs, inlineRun{text: t, bold: bold, italic: italic, mono: mono})
			}
		case *ast.Emphasis:
			// Level 1 = italic, Level 2 = bold (strong)
			childBold := bold || node.Level == 2
			childItalic := italic || node.Level == 1
			runs = append(runs, w.buildRuns(node, childBold, childItalic, mono)...)
		case *ast.CodeSpan:
			var sb strings.Builder
			collectTextInto(node, w.src, &sb)
			runs = append(runs, inlineRun{text: sb.String(), mono: true})
		case *ast.Link:
			runs = append(runs, w.buildRuns(node, bold, italic, mono)...)
		default:
			runs = append(runs, w.buildRuns(child, bold, italic, mono)...)
		}
	}
	return runs
}

func runsToString(runs []inlineRun) string {
	var sb strings.Builder
	for _, r := range runs {
		sb.WriteString(r.text)
	}
	return sb.String()
}
