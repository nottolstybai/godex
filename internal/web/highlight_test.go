package web

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHighlightGo(t *testing.T) {
	t.Parallel()

	const (
		kw  = `<span class="tok-kw">`
		str = `<span class="tok-str">`
		num = `<span class="tok-num">`
		com = `<span class="tok-com">`
		typ = `<span class="tok-type">`
		fn  = `<span class="tok-fn">`
		end = `</span>`
	)

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"empty", "", ""},
		{"function", "func Do(s string) error", kw + "func" + end + " " + fn + "Do" + end + "(s " + typ + "string" + end + ") " + typ + "error" + end},
		{"exported type", "var r io.Reader", kw + "var" + end + " r io." + typ + "Reader" + end},
		{"call", "client.Do(req)", "client." + fn + "Do" + end + "(req)"},
		{"number", "x = 1.5", "x = " + num + "1.5" + end},
		{"nil", "return nil", kw + "return" + end + " " + num + "nil" + end},
		{"comment", "// note\nx", com + "// note" + end + "\nx"},
		{"keeps whitespace", "x\n\ty  z", "x\n\ty  z"},
		{"escapes string", `"<b>"`, str + "&#34;&lt;b&gt;&#34;" + end},
		{"escapes operators", "a <- b", "a &lt;- b"},
		{"escapes comment", "// <script>", com + "// &lt;script&gt;" + end},
		{"not Go", "$ <foo> run", "$ &lt;foo&gt; run"},
		{"unterminated string", `"abc`, str + `&#34;abc` + end},
		{"invalid UTF-8", "\xff<", "\xff&lt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, string(highlightGo(tt.src)))
		})
	}

	t.Run("output keeps the source text", func(t *testing.T) {
		t.Parallel()

		src := "type T struct {\n\tA int // first\n\tB string\n}\n\nfunc (t *T) M(x []byte) (n int, err error) {\n\treturn len(x), nil\n}\n"

		got := string(highlightGo(src))
		got = tagPattern.ReplaceAllString(got, "")

		assert.Equal(t, src, strings.NewReplacer("&lt;", "<", "&gt;", ">", "&#34;", `"`).Replace(got))
	})
}
