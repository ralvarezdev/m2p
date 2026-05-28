package converter

import (
	"context"
	"fmt"
	"strings"
)

// Engine selects the PDF rendering backend.
type Engine string

const (
	// EngineAuto tries Chromium first and falls back to Native if no browser is found.
	EngineAuto     Engine = "auto"
	EngineChromium Engine = "chromium"
	EngineNative   Engine = "native"
)

func (e Engine) String() string { return string(e) }

func ParseEngine(s string) (Engine, error) {
	switch strings.ToLower(s) {
	case "auto":
		return EngineAuto, nil
	case "chromium":
		return EngineChromium, nil
	case "native":
		return EngineNative, nil
	default:
		return "", fmt.Errorf("unknown engine %q: must be auto, chromium, or native", s)
	}
}

// RenderInput is the structured document passed to every PDF engine.
type RenderInput struct {
	// Source is the raw Markdown bytes — each engine parses it independently.
	Source     []byte
	Title      string
	Date       string
	ShowFooter bool
	Paper      Paper
	Output     string
}

// Renderer converts a Markdown document to a PDF file.
type Renderer interface {
	Render(ctx context.Context, in RenderInput) error
}

// NewRenderer returns a Renderer for the given engine. EngineAuto returns a
// chromiumRenderer that transparently falls back to nativeRenderer when no
// browser is found.
func NewRenderer(engine Engine) Renderer {
	switch engine {
	case EngineNative:
		return &nativeRenderer{}
	case EngineChromium:
		return &chromiumRenderer{requireBrowser: true}
	default: // EngineAuto
		return &chromiumRenderer{requireBrowser: false}
	}
}
