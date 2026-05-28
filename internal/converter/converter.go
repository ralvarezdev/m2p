package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Format string

const (
	FormatPDF  Format = "pdf"
	FormatHTML Format = "html"
	FormatBoth Format = "both"
)

func (f Format) String() string { return string(f) }

func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case formatPDFStr:
		return FormatPDF, nil
	case formatHTMLStr:
		return FormatHTML, nil
	case formatBothStr:
		return FormatBoth, nil
	default:
		return "", fmt.Errorf("unknown format %q: must be pdf, html, or both", s)
	}
}

type Paper string

const (
	PaperA4     Paper = "a4"
	PaperLetter Paper = "letter"
	PaperA3     Paper = "a3"
	PaperLegal  Paper = "legal"
)

func (p Paper) String() string { return string(p) }

func ParsePaper(s string) (Paper, error) {
	switch strings.ToLower(s) {
	case paperA4Str:
		return PaperA4, nil
	case paperLetterStr:
		return PaperLetter, nil
	case paperA3Str:
		return PaperA3, nil
	case paperLegalStr:
		return PaperLegal, nil
	default:
		return "", fmt.Errorf("unknown paper size %q: must be a4, letter, a3, or legal", s)
	}
}

// Options configures a single conversion run.
type Options struct {
	Input          string
	Output         string
	Format         Format
	Paper          Paper
	Engine         Engine
	ShowFooter     bool
	Open           bool
	PageBreakLevel int // 0 = none, 2 = h2, 3 = h3
}

// ParsePageBreak converts the --page-break flag value to a heading level int.
func ParsePageBreak(s string) (int, error) {
	switch strings.ToLower(s) {
	case pageBreakNoneStr, "":
		return 0, nil
	case pageBreakH2Str:
		return 2, nil
	case pageBreakH3Str:
		return 3, nil
	default:
		return 0, fmt.Errorf("unknown page-break value %q: must be none, h2, or h3", s)
	}
}

// DefaultOutput returns the output path derived from input when not specified.
func DefaultOutput(input string, format Format) string {
	ext := ExtPDF
	if format == FormatHTML {
		ext = ExtHTML
	}
	base := strings.TrimSuffix(input, filepath.Ext(input))
	return base + ext
}

// Convert runs the full pipeline for the given options.
func Convert(opts Options) error {
	src, err := os.ReadFile(opts.Input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	title := strings.TrimSuffix(filepath.Base(opts.Input), filepath.Ext(opts.Input))
	date := time.Now().Format(DateFormat)

	switch opts.Format {
	case FormatHTML:
		return convertHTML(src, opts, title, date)
	case FormatPDF:
		return convertPDF(src, opts, title, date)
	case FormatBoth:
		if err := convertHTML(src, opts, title, date); err != nil {
			return err
		}
		return convertPDF(src, opts, title, date)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

func convertHTML(src []byte, opts Options, title, date string) error {
	fragment, mdTitle, err := ParseMarkdown(src)
	if err != nil {
		return fmt.Errorf("parse markdown: %w", err)
	}
	if mdTitle != "" {
		title = mdTitle
	}
	htmlBytes, err := RenderTemplate(fragment, title, date, opts.ShowFooter, opts.PageBreakLevel)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	htmlOut := strings.TrimSuffix(opts.Output, filepath.Ext(opts.Output)) + ExtHTML
	return writeFile(htmlOut, htmlBytes)
}

func convertPDF(src []byte, opts Options, title, date string) error {
	pdfOut := strings.TrimSuffix(opts.Output, filepath.Ext(opts.Output)) + ExtPDF

	renderer := NewRenderer(opts.Engine)
	return renderer.Render(context.Background(), RenderInput{
		Source:         src,
		Title:          title,
		Date:           date,
		ShowFooter:     opts.ShowFooter,
		Paper:          opts.Paper,
		Output:         pdfOut,
		PageBreakLevel: opts.PageBreakLevel,
	})
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), DirPerm); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return os.WriteFile(path, data, FilePerm)
}
