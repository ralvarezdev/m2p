package converter

// File extensions
const (
	ExtMarkdown = ".md"
	ExtPDF      = ".pdf"
	ExtHTML     = ".html"
)

// Date format for document metadata
const DateFormat = "2006-01-02"

// File permission constants
const (
	DirPerm  = 0o755
	FilePerm = 0o644
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
