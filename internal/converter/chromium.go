package converter

import (
	"context"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// chromeCandidates returns the ordered list of Chrome/Chromium executable paths
// for the current OS. Evaluated at call time so LOCALAPPDATA is read after
// program startup rather than at package init.
func chromeCandidates() []string {
	switch runtime.GOOS {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		return []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			filepath.Join(local, "Google", "Chrome", "Application", "chrome.exe"),
			`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
			filepath.Join(local, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Chromium\Application\chrome.exe`,
		}
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	default:
		return []string{
			"google-chrome",
			"google-chrome-stable",
			"brave-browser",
			"microsoft-edge",
			"chromium",
			"chromium-browser",
		}
	}
}

// findChrome returns the path to a usable Chrome/Chromium browser, or an error
// with install instructions when none is found.
func findChrome() (string, error) {
	candidates := chromeCandidates()
	for _, c := range candidates {
		if filepath.IsAbs(c) {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		} else {
			if p, err := exec.LookPath(c); err == nil {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf(
		"no Chromium-based browser found — install Chrome, Brave, or Edge, or use --engine native",
	)
}

// getPaperParams maps a paper size to chromedp print params. When showFooter is
// true it activates Chrome's native footer template, which renders in the bottom
// margin area and never overlaps body content.
func getPaperParams(p Paper, showFooter bool, date string) page.PrintToPDFParams {
	dims := PaperSizes[p]
	params := page.PrintToPDFParams{
		PrintBackground: true,
		PaperWidth:      dims.WidthInches,
		PaperHeight:     dims.HeightInches,
		MarginTop:       dims.MarginInches,
		MarginBottom:    dims.MarginInches,
		MarginLeft:      dims.MarginInches,
		MarginRight:     dims.MarginInches,
	}
	if showFooter {
		params.DisplayHeaderFooter = true
		params.HeaderTemplate = ChromeHeaderPlaceholder
		params.FooterTemplate = buildNativeFooterTemplate(date, dims.MarginInches)
	}
	return params
}

// buildNativeFooterTemplate returns the HTML injected into Chrome's bottom margin
// on every page. Inline styles are required; CSS custom properties are unavailable.
func buildNativeFooterTemplate(date string, marginInches float64) string {
	paddingPx := marginInches * ScreenDPI
	return fmt.Sprintf(
		`<div style="font-family:%s;font-size:%dpx;color:%s;width:100%%;`+
			`display:flex;justify-content:space-between;align-items:center;`+
			`padding:0 %.1fpx;box-sizing:border-box;">`+
			`<span>0x<span style="color:%s">RA</span>&nbsp;&middot;&nbsp;m2p</span>`+
			`<span>%s</span>`+
			`</div>`,
		ChromeFooterFontFamily, ChromeFooterFontSizePx, ChromeFooterColorMuted, paddingPx,
		ChromeFooterColorCyan, html.EscapeString(date),
	)
}

type chromiumRenderer struct {
	// requireBrowser causes Render to return an error when no browser is found
	// instead of falling back to the native renderer.
	requireBrowser bool
}

func (r *chromiumRenderer) Render(ctx context.Context, in RenderInput) error {
	chromeBin, err := findChrome()
	if err != nil {
		if r.requireBrowser {
			return err
		}
		// Auto mode: fall back gracefully to native renderer.
		native := &nativeRenderer{}
		return native.Render(ctx, in)
	}

	fragment, title, err := ParseMarkdown(in.Source)
	if err != nil {
		return fmt.Errorf("parse markdown: %w", err)
	}
	if in.Title == "" {
		in.Title = title
	}

	htmlBytes, err := RenderTemplate(fragment, in.Title, in.Date, in.ShowFooter, in.PageBreakLevel)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return renderWithChrome(ctx, chromeBin, htmlBytes, in.Output, in.Paper, in.ShowFooter, in.Date)
}

func renderWithChrome(
	ctx context.Context, chromeBin string, htmlBytes []byte,
	output string, paper Paper, showFooter bool, date string,
) error {
	if _, ok := PaperSizes[paper]; !ok {
		return fmt.Errorf("unknown paper size: %s", paper)
	}
	params := getPaperParams(paper, showFooter, date)

	tmp, err := os.CreateTemp("", "m2p-*.html")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(htmlBytes); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	tmpURL := "file:///" + filepath.ToSlash(tmp.Name())

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(
		ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(chromeBin),
		)...,
	)
	defer cancelAlloc()

	chromCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfData []byte
	if err := chromedp.Run(chromCtx,
		chromedp.Navigate(tmpURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfData, _, err = page.PrintToPDF().
				WithPrintBackground(params.PrintBackground).
				WithPaperWidth(params.PaperWidth).
				WithPaperHeight(params.PaperHeight).
				WithMarginTop(params.MarginTop).
				WithMarginBottom(params.MarginBottom).
				WithMarginLeft(params.MarginLeft).
				WithMarginRight(params.MarginRight).
				WithDisplayHeaderFooter(params.DisplayHeaderFooter).
				WithHeaderTemplate(params.HeaderTemplate).
				WithFooterTemplate(params.FooterTemplate).
				Do(ctx)
			return err
		}),
	); err != nil {
		return fmt.Errorf("chromedp: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(output), DirPerm); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return os.WriteFile(output, pdfData, FilePerm)
}
