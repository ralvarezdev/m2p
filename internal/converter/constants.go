package converter

// File extensions
const (
	ExtMarkdown = ".md"
	ExtPDF      = ".pdf"
	ExtHTML     = ".html"
)

// DateFormat is the date format for document metadata
const DateFormat = "2006-01-02"

// File permission constants
const (
	DirPerm  = 0o755
	FilePerm = 0o644
)

// PaperDimensions defines width, height, and margins for a paper type.
type PaperDimensions struct {
	WidthInches  float64
	HeightInches float64
	MarginInches float64
}

// PaperSizes maps each paper type to its dimensions (in inches for chromium, mm for native).
var PaperSizes = map[Paper]PaperDimensions{
	PaperA4: {
		WidthInches:  8.27,
		HeightInches: 11.69,
		MarginInches: 0.6,
	},
	PaperLetter: {
		WidthInches:  8.5,
		HeightInches: 11,
		MarginInches: 0.6,
	},
	PaperA3: {
		WidthInches:  11.69,
		HeightInches: 16.54,
		MarginInches: 0.6,
	},
	PaperLegal: {
		WidthInches:  8.5,
		HeightInches: 14,
		MarginInches: 0.6,
	},
}

// NativeMargin represents PDF margins in millimeters for the native renderer.
type NativeMargin int

const (
	NativeMarginDefault NativeMargin = 20 // Top, left, right margins
	NativeMarginFooter  NativeMargin = 30 // Bottom margin when footer is shown
)

// TypographySize represents font sizes in points for document elements.
type TypographySize int

const (
	TypoH1   TypographySize = 26 // Title/document heading
	TypoH2   TypographySize = 22 // H2 heading
	TypoH3   TypographySize = 16 // H3 heading
	TypoH4   TypographySize = 14 // H4 heading
	TypoH5   TypographySize = 12 // H5 heading
	TypoH6   TypographySize = 11 // H6 heading
	TypoBody TypographySize = 11 // Body text / default
	TypoCode TypographySize = 10 // Code blocks & blockquotes
)

// CodeBlockDimension represents layout measurements for code blocks and blockquotes.
type CodeBlockDimension float64

const (
	CodeBlockLineHeight CodeBlockDimension = 4.5 // Height per line in code box
	CodeBlockPaddingX   CodeBlockDimension = 16  // Horizontal padding
	CodeBlockPaddingY   CodeBlockDimension = 10  // Vertical padding
	CodeBlockHeaderH    CodeBlockDimension = 4   // Decorative header height
	CodeBlockHeaderW    CodeBlockDimension = 18  // Decorative header width
	CodeBlockMarginR    CodeBlockDimension = 20  // Right margin for header
	CodeBlockMarginT    CodeBlockDimension = 2   // Top margin for header
)

// HighlightColor represents RGB component values for syntax highlighting fallback.
type HighlightColor uint8

const (
	HighlightColorR HighlightColor = 192 // Red component (#c0caf5)
	HighlightColorG HighlightColor = 202 // Green component (#c0caf5)
	HighlightColorB HighlightColor = 245 // Blue component (#c0caf5)
)

// String representations of typed constants for parsing
const (
	formatPDFStr  = "pdf"
	formatHTMLStr = "html"
	formatBothStr = "both"

	paperA4Str     = "a4"
	paperLetterStr = "letter"
	paperA3Str     = "a3"
	paperLegalStr  = "legal"

	engineAutoStr     = "auto"
	engineChromiumStr = "chromium"
	engineNativeStr   = "native"

	pageBreakNoneStr = "none"
	pageBreakH2Str   = "h2"
	pageBreakH3Str   = "h3"
)

// Font CSS templates with base64 data URI placeholders
const (
	interFontTemplate = `@font-face {
  font-family: 'Inter';
  font-style: normal;
  font-weight: 100 900;
  font-display: swap;
  src: url(data:font/woff2;base64,%s) format('woff2');
  unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA,
    U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191,
    U+2193, U+2212, U+2215, U+FEFF, U+FFFD;
}`

	jetBrainsMonoTemplate = `@font-face {
  font-family: 'JetBrains Mono';
  font-style: normal;
  font-weight: 400 600;
  font-display: swap;
  src: url(data:font/woff2;base64,%s) format('woff2');
  unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA,
    U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191,
    U+2193, U+2212, U+2215, U+FEFF, U+FFFD;
}`
)
