package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want []string
	}{
		{"lowercase", "Read FILE", []string{"read", "file"}},
		{"punctuation", "Read, until EOF.", []string{"read", "until", "eof"}},
		{"dot in name", "io.Reader", []string{"io", "reader"}},
		{"digits", "HTTP/2 server", []string{"http", "2", "server"}},
		{"newlines", "first line\nsecond line", []string{"first", "line", "second", "line"}},
		{"CamelCase", "ReadAll", []string{"readall", "read", "all"}},
		{"three-word CamelCase", "CloseIdleConnections", []string{"closeidleconnections", "close", "idle", "connections"}},
		{"leading acronym", "HTTPServer", []string{"httpserver", "http", "server"}},
		{"trailing acronym", "parseURL", []string{"parseurl", "parse", "url"}},
		{"trailing digits", "Sum256", []string{"sum256"}},
		{"digits in the middle", "Base64Encoding", []string{"base64encoding", "base64", "encoding"}},
		{"identifier in a sentence", "Use io.ReadAll now", []string{"use", "io", "readall", "read", "all", "now"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, Tokenize(tt.text))
		})
	}
}

func TestTokenizeEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
	}{
		{"empty string", ""},
		{"separators only", " ,.!\n\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, Tokenize(tt.text))
		})
	}
}
