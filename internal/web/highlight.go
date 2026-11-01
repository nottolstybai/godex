package web

import (
	"go/scanner"
	"go/token"
	"html"
	"html/template"
	"strings"
	"unicode"
)

var builtinTypes = map[string]bool{
	"any": true, "bool": true, "byte": true, "comparable": true, "complex64": true, "complex128": true,
	"error": true, "float32": true, "float64": true, "int": true, "int8": true, "int16": true, "int32": true,
	"int64": true, "rune": true, "string": true, "uint": true, "uint8": true, "uint16": true, "uint32": true,
	"uint64": true, "uintptr": true,
}

type lexeme struct {
	offset int
	tok    token.Token
	text   string
}

// highlightGo escapes src and wraps Go tokens in <span> elements.
// Text that isn't valid Go is still escaped, just with less color.
func highlightGo(src string) template.HTML {
	lexemes := scanGo(src)

	var out strings.Builder

	last := 0
	for i, lex := range lexemes {
		end := min(lex.offset+len(lex.text), len(src))
		if lex.offset < last || lex.offset >= end {
			continue
		}

		out.WriteString(html.EscapeString(src[last:lex.offset]))

		followedByParen := i+1 < len(lexemes) && lexemes[i+1].tok == token.LPAREN
		text := html.EscapeString(src[lex.offset:end])
		if class := tokenClass(lex, followedByParen); class != "" {
			out.WriteString(`<span class="`)
			out.WriteString(class)
			out.WriteString(`">`)
			out.WriteString(text)
			out.WriteString(`</span>`)
		} else {
			out.WriteString(text)
		}

		last = end
	}
	out.WriteString(html.EscapeString(src[last:]))

	return template.HTML(out.String()) //nolint:gosec // escaped above
}

func scanGo(src string) []lexeme {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))

	var s scanner.Scanner
	s.Init(file, []byte(src), nil, scanner.ScanComments)

	var lexemes []lexeme
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			return lexemes
		}
		// Semicolons inserted by the scanner aren't in the source.
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}

		text := lit
		if text == "" {
			text = tok.String()
		}
		lexemes = append(lexemes, lexeme{offset: file.Offset(pos), tok: tok, text: text})
	}
}

func tokenClass(lex lexeme, followedByParen bool) string {
	switch {
	case lex.tok.IsKeyword():
		return "tok-kw"
	case lex.tok == token.STRING, lex.tok == token.CHAR:
		return "tok-str"
	case lex.tok == token.INT, lex.tok == token.FLOAT, lex.tok == token.IMAG:
		return "tok-num"
	case lex.tok == token.COMMENT:
		return "tok-com"
	case lex.tok != token.IDENT:
		return ""
	case lex.text == "nil", lex.text == "true", lex.text == "false":
		return "tok-num"
	case followedByParen:
		return "tok-fn"
	case builtinTypes[lex.text], unicode.IsUpper([]rune(lex.text)[0]):
		return "tok-type"
	default:
		return ""
	}
}
