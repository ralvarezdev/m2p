package assets

import _ "embed"

var (
	//go:embed templates/document.html
	DocumentHTML string

	//go:embed styles/styles.css
	StylesCSS string

	//go:embed fonts/inter-latin.woff2
	InterFont []byte

	//go:embed fonts/jetbrains-mono-latin.woff2
	JetBrainsMonoFont []byte
)
