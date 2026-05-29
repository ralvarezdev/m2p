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
		chroma.Background: tokyoNightFg + " bg:#1a1b2e",
		chroma.Text:       tokyoNightFg,
		chroma.Other:      tokyoNightFg,

		// Comments — muted blue-grey
		chroma.Comment:          tokyoNightComment,
		chroma.CommentHashbang:  tokyoNightComment,
		chroma.CommentMultiline: tokyoNightComment,
		chroma.CommentSingle:    tokyoNightComment,
		chroma.CommentSpecial:   tokyoNightComment,
		chroma.CommentPreproc:   tokyoNightPurple,

		// Keywords — purple
		chroma.Keyword:            tokyoNightPurple,
		chroma.KeywordConstant:    tokyoNightPurple,
		chroma.KeywordDeclaration: tokyoNightPurple,
		chroma.KeywordNamespace:   tokyoNightPurple,
		chroma.KeywordPseudo:      tokyoNightPurple,
		chroma.KeywordReserved:    tokyoNightPurple,
		chroma.KeywordType:        ChromeFooterColorCyan, // primitive types (int, string…) — cyan

		// Names
		chroma.Name:                 tokyoNightFg,
		chroma.NameAttribute:        tokyoNightBlue,        // struct tags / HTML attrs — blue
		chroma.NameBuiltin:          ChromeFooterColorCyan, // built-in functions — cyan
		chroma.NameBuiltinPseudo:    tokyoNightFg,
		chroma.NameClass:            ChromeFooterColorCyan, // type / class names — cyan
		chroma.NameConstant:         tokyoNightOrange,      // named constants — orange
		chroma.NameDecorator:        tokyoNightBlue,
		chroma.NameEntity:           tokyoNightFg,
		chroma.NameException:        ChromeFooterColorCyan,
		chroma.NameFunction:         tokyoNightBlue, // function names — blue
		chroma.NameLabel:            ChromeFooterColorCyan + " italic",
		chroma.NameNamespace:        tokyoNightFg,
		chroma.NameOther:            tokyoNightFg,
		chroma.NameTag:              tokyoNightPurple,
		chroma.NameVariable:         tokyoNightFg, // local variables — neutral fg
		chroma.NameVariableClass:    ChromeFooterColorCyan,
		chroma.NameVariableGlobal:   tokyoNightFg,
		chroma.NameVariableInstance: "#e0af68", // instance fields — warm yellow

		// Literals
		chroma.Literal:     tokyoNightFg,
		chroma.LiteralDate: tokyoNightGreen,

		// Strings — green
		chroma.LiteralString:         tokyoNightGreen,
		chroma.LiteralStringBacktick: tokyoNightGreen,
		chroma.LiteralStringChar:     tokyoNightGreen,
		chroma.LiteralStringDoc:      tokyoNightGreen,
		chroma.LiteralStringDouble:   tokyoNightGreen,
		chroma.LiteralStringEscape:   tokyoNightPurple, // escape sequences — purple
		chroma.LiteralStringHeredoc:  tokyoNightGreen,
		chroma.LiteralStringInterpol: tokyoNightGreen,
		chroma.LiteralStringOther:    tokyoNightGreen,
		chroma.LiteralStringRegex:    tokyoNightGreen,
		chroma.LiteralStringSingle:   tokyoNightGreen,
		chroma.LiteralStringSymbol:   tokyoNightGreen,

		// Numbers — orange
		chroma.LiteralNumber:            tokyoNightOrange,
		chroma.LiteralNumberBin:         tokyoNightOrange,
		chroma.LiteralNumberFloat:       tokyoNightOrange,
		chroma.LiteralNumberHex:         tokyoNightOrange,
		chroma.LiteralNumberInteger:     tokyoNightOrange,
		chroma.LiteralNumberIntegerLong: tokyoNightOrange,
		chroma.LiteralNumberOct:         tokyoNightOrange,

		// Operators & punctuation
		chroma.Operator:     "#89ddff",        // light cyan
		chroma.OperatorWord: tokyoNightPurple, // word operators — purple
		chroma.Punctuation:  tokyoNightFg,

		// Generics (diffs, shell output…)
		chroma.Generic:           tokyoNightFg,
		chroma.GenericDeleted:    tokyoNightRed,
		chroma.GenericEmph:       "italic",
		chroma.GenericError:      tokyoNightRed,
		chroma.GenericHeading:    tokyoNightBlue + " bold",
		chroma.GenericInserted:   tokyoNightGreen + " bold",
		chroma.GenericOutput:     tokyoNightComment,
		chroma.GenericPrompt:     tokyoNightFg,
		chroma.GenericStrong:     "bold",
		chroma.GenericSubheading: tokyoNightBlue + " bold",
		chroma.GenericTraceback:  tokyoNightFg,
		chroma.GenericUnderline:  "underline",

		chroma.Error: tokyoNightRed,
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

	iter, _ := lexer.Tokenise(nil, src)
	if iter != nil {
		return iter.Tokens(), nil
	}
	return []chroma.Token{{Type: chroma.Text, Value: src}}, nil
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
	var current []chroma.Token

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
