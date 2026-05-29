package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// chromePaths lists candidate Chrome/Chromium executables per OS, in order of preference.
var chromePaths = map[string][]string{
	"windows": {
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Google\Chrome\Application\chrome.exe`),
		`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `BraveSoftware\Brave-Browser\Application\brave.exe`),
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Chromium\Application\chrome.exe`,
	},
	"darwin": {
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	},
	"linux": {
		"google-chrome",
		"google-chrome-stable",
		"brave-browser",
		"microsoft-edge",
		"chromium",
		"chromium-browser",
	},
}

// findChrome returns the path to a usable Chrome/Chromium browser, or an error
// with install instructions when none is found.
func findChrome() (string, error) {
	candidates := chromePaths[runtime.GOOS]
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

// paperParams maps paper size names to chromedp print params (dimensions in inches).
func getPaperParams(p Paper) page.PrintToPDFParams {
	dims := PaperSizes[p]
	return page.PrintToPDFParams{
		PrintBackground: true,
		PaperWidth:      dims.WidthInches,
		PaperHeight:     dims.HeightInches,
		MarginTop:       dims.MarginInches,
		MarginBottom:    dims.MarginInches,
		MarginLeft:      dims.MarginInches,
		MarginRight:     dims.MarginInches,
	}
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

	return renderWithChrome(ctx, chromeBin, htmlBytes, in.Output, in.Paper)
}

func renderWithChrome(_ context.Context, chromeBin string, htmlBytes []byte, output string, paper Paper) error {
	if _, ok := PaperSizes[paper]; !ok {
		return fmt.Errorf("unknown paper size: %s", paper)
	}
	params := getPaperParams(paper)

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
		context.Background(),
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
				Do(ctx)
			return err
		}),
	); err != nil {
		return fmt.Errorf("chromedp: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return os.WriteFile(output, pdfData, 0o644)
}
