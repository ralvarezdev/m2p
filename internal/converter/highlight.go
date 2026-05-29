package converter

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// highlightTheme is the chroma style used by both rendering engines.
const highlightTheme = "tokyonight-dark"

func init() {
	// Register a custom Tokyo Night Dark style since chroma v2.2.0 does not
	// bundle it. Token type names are constrained to what v2.2.0 exports.
	//
	// Key distinctions:
	//   Variables (NameVariable*)  → neutral fg  #c0caf5
	//   Types/classes (NameClass)  → cyan        #7dcfff
	//   Functions (NameFunction)   → blue        #7aa2f7
	//   Keywords                   → purple      #bb9af7
	styles.Register(chroma.MustNewStyle("tokyonight-dark", chroma.StyleEntries{
		chroma.Background: "#c0caf5 bg:#1a1b2e",
		chroma.Text:       "#c0caf5",
		chroma.Other:      "#c0caf5",

		// Comments — muted blue-grey
		chroma.Comment:          "#565f89",
		chroma.CommentHashbang:  "#565f89",
		chroma.CommentMultiline: "#565f89",
		chroma.CommentSingle:    "#565f89",
		chroma.CommentSpecial:   "#565f89",
		chroma.CommentPreproc:   "#bb9af7",

		// Keywords — purple
		chroma.Keyword:            "#bb9af7",
		chroma.KeywordConstant:    "#bb9af7",
		chroma.KeywordDeclaration: "#bb9af7",
		chroma.KeywordNamespace:   "#bb9af7",
		chroma.KeywordPseudo:      "#bb9af7",
		chroma.KeywordReserved:    "#bb9af7",
		chroma.KeywordType:        "#7dcfff", // primitive types (int, string…) — cyan

		// Names
		chroma.Name:                 "#c0caf5",
		chroma.NameAttribute:        "#7aa2f7",           // struct tags / HTML attrs — blue
		chroma.NameBuiltin:          "#7dcfff",           // built-in functions — cyan
		chroma.NameBuiltinPseudo:    "#c0caf5",
		chroma.NameClass:            "#7dcfff",           // type / class names — cyan
		chroma.NameConstant:         "#ff9e64",           // named constants — orange
		chroma.NameDecorator:        "#7aa2f7",
		chroma.NameEntity:           "#c0caf5",
		chroma.NameException:        "#7dcfff",
		chroma.NameFunction:         "#7aa2f7",           // function names — blue
		chroma.NameLabel:            "#7dcfff italic",
		chroma.NameNamespace:        "#c0caf5",
		chroma.NameOther:            "#c0caf5",
		chroma.NameTag:              "#bb9af7",
		chroma.NameVariable:         "#c0caf5",           // local variables — neutral fg
		chroma.NameVariableClass:    "#7dcfff",
		chroma.NameVariableGlobal:   "#c0caf5",
		chroma.NameVariableInstance: "#e0af68",           // instance fields — warm yellow

		// Literals
		chroma.Literal:     "#c0caf5",
		chroma.LiteralDate: "#9ece6a",

		// Strings — green
		chroma.LiteralString:         "#9ece6a",
		chroma.LiteralStringBacktick: "#9ece6a",
		chroma.LiteralStringChar:     "#9ece6a",
		chroma.LiteralStringDoc:      "#9ece6a",
		chroma.LiteralStringDouble:   "#9ece6a",
		chroma.LiteralStringEscape:   "#bb9af7", // escape sequences — purple
		chroma.LiteralStringHeredoc:  "#9ece6a",
		chroma.LiteralStringInterpol: "#9ece6a",
		chroma.LiteralStringOther:    "#9ece6a",
		chroma.LiteralStringRegex:    "#9ece6a",
		chroma.LiteralStringSingle:   "#9ece6a",
		chroma.LiteralStringSymbol:   "#9ece6a",

		// Numbers — orange
		chroma.LiteralNumber:            "#ff9e64",
		chroma.LiteralNumberBin:         "#ff9e64",
		chroma.LiteralNumberFloat:       "#ff9e64",
		chroma.LiteralNumberHex:         "#ff9e64",
		chroma.LiteralNumberInteger:     "#ff9e64",
		chroma.LiteralNumberIntegerLong: "#ff9e64",
		chroma.LiteralNumberOct:         "#ff9e64",

		// Operators & punctuation
		chroma.Operator:     "#89ddff", // light cyan
		chroma.OperatorWord: "#bb9af7", // word operators — purple
		chroma.Punctuation:  "#c0caf5",

		// Generics (diffs, shell output…)
		chroma.Generic:           "#c0caf5",
		chroma.GenericDeleted:    "#f7768e",
		chroma.GenericEmph:       "italic",
		chroma.GenericError:      "#f7768e",
		chroma.GenericHeading:    "#7aa2f7 bold",
		chroma.GenericInserted:   "#9ece6a bold",
		chroma.GenericOutput:     "#565f89",
		chroma.GenericPrompt:     "#c0caf5",
		chroma.GenericStrong:     "bold",
		chroma.GenericSubheading: "#7aa2f7 bold",
		chroma.GenericTraceback:  "#c0caf5",
		chroma.GenericUnderline:  "underline",

		chroma.Error: "#f7768e",
	}))
}

// tokenize returns the chroma token stream for the given language and source.
// Falls back to plain-text when the language is unknown.
func tokenize(lang, src string) ([]chroma.Token, error) {
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(strings.ToLower(lang))
	}
	if lexer == nil {
		lexer = lexers.Analyse(src)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iter, err := lexer.Tokenise(nil, src)
	if err != nil {
		return []chroma.Token{{Type: chroma.Text, Value: src}}, nil
	}
	return iter.Tokens(), nil
}

// tokenRGB returns the foreground RGB for a chroma token type using highlightTheme.
// Falls back to plain code foreground (#c0caf5) when the token has no assigned color.
func tokenRGB(t chroma.TokenType) (r, g, b uint8) {
	style := styles.Get(highlightTheme)
	if style == nil {
		return uint8(HighlightColorR), uint8(HighlightColorG), uint8(HighlightColorB)
	}
	if c := style.Get(t).Colour; c != 0 {
		return c.Red(), c.Green(), c.Blue()
	}
	return uint8(HighlightColorR), uint8(HighlightColorG), uint8(HighlightColorB)
}

// splitTokensByLine groups a flat token list into per-line slices, splitting
// on \n within token values.
func splitTokensByLine(tokens []chroma.Token) [][]chroma.Token {
	var lines [][]chroma.Token
	current := []chroma.Token{}

	for _, tok := range tokens {
		parts := strings.Split(tok.Value, "\n")
		for i, part := range parts {
			if part != "" {
				current = append(current, chroma.Token{Type: tok.Type, Value: part})
			}
			if i < len(parts)-1 {
				lines = append(lines, current)
				current = []chroma.Token{}
			}
		}
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}
	return lines
}
