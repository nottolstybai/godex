package search

import (
	"strings"
	"unicode"
)

func Tokenize(text string) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	})

	tokens := make([]string, 0, len(words))
	for _, word := range words {
		tokens = append(tokens, strings.ToLower(word))

		parts := splitCamelCase(word)
		if len(parts) < 2 {
			continue
		}
		for _, part := range parts {
			tokens = append(tokens, strings.ToLower(part))
		}
	}

	return tokens
}

// splitCamelCase splits camel case word: HTTPServer → HTTP, Server
func splitCamelCase(word string) []string {
	runes := []rune(word)

	var parts []string
	start := 0
	for i := 1; i < len(runes); i++ {
		prev, cur := runes[i-1], runes[i]

		afterLowerOrDigit := (unicode.IsLower(prev) || unicode.IsDigit(prev)) && unicode.IsUpper(cur)
		endOfAcronym := unicode.IsUpper(prev) && unicode.IsUpper(cur) && i+1 < len(runes) && unicode.IsLower(runes[i+1])

		if afterLowerOrDigit || endOfAcronym {
			parts = append(parts, string(runes[start:i]))
			start = i
		}
	}

	return append(parts, string(runes[start:]))
}
