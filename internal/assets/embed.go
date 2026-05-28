package assets

import _ "embed"

//go:embed templates/document.html
var DocumentHTML string

//go:embed styles/styles.css
var StylesCSS string

//go:embed fonts/inter-latin.woff2
var InterFont []byte

//go:embed fonts/jetbrains-mono-latin.woff2
var JetBrainsMonoFont []byte
