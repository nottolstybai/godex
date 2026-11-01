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
		{"нижний регистр", "Read FILE", []string{"read", "file"}},
		{"пунктуация", "Read, until EOF.", []string{"read", "until", "eof"}},
		{"точка в имени", "io.Reader", []string{"io", "reader"}},
		{"цифры", "HTTP/2 server", []string{"http", "2", "server"}},
		{"переводы строк", "first line\nsecond line", []string{"first", "line", "second", "line"}},
		{"CamelCase", "ReadAll", []string{"readall", "read", "all"}},
		{"CamelCase из трёх слов", "CloseIdleConnections", []string{"closeidleconnections", "close", "idle", "connections"}},
		{"аббревиатура в начале", "HTTPServer", []string{"httpserver", "http", "server"}},
		{"аббревиатура в конце", "parseURL", []string{"parseurl", "parse", "url"}},
		{"цифры в конце слова", "Sum256", []string{"sum256"}},
		{"цифры в середине", "Base64Encoding", []string{"base64encoding", "base64", "encoding"}},
		{"идентификатор в предложении", "Use io.ReadAll now", []string{"use", "io", "readall", "read", "all", "now"}},
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
		{"пустая строка", ""},
		{"только разделители", " ,.!\n\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, Tokenize(tt.text))
		})
	}
}
