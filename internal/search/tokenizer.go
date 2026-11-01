package search

import (
	"strings"
	"unicode"
)

func Tokenize(text string) []string {
	text = strings.ToLower(text)
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	})
}
