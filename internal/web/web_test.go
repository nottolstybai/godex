package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nottolstybai/godex/internal/search"
)

func testDocs() []search.Document {
	return []search.Document{
		{Name: "os.ReadFile", Package: "os", Decl: "func ReadFile(name string) ([]byte, error)", Text: "ReadFile reads the named file and returns the contents."},
		{Name: "os.File", Package: "os", Decl: "type File struct{}", Text: "File represents an open file descriptor."},
		{Name: "net/http", Package: "net/http", Text: "Package http provides HTTP client and server implementations."},
		{Name: "net/http.Client.Do", Package: "net/http", Decl: "func (c *Client) Do(req *Request) (*Response, error)", Text: "Do sends an HTTP request. See [Client] and [io.Reader].\n\nExample:\n\n\tresp, err := client.Do(req)"},
		{Name: "a/internal/b.Hidden", Package: "a/internal/b", Decl: "func Hidden()", Text: "Hidden is an internal function."},
		{Name: "x.Evil", Package: "x", Decl: "func Evil(c <-chan int)", Text: "Evil prints <script>alert(1)</script> on the page."},
	}
}

var tagPattern = regexp.MustCompile(`<[^>]+>`)

func get(t *testing.T, handler http.Handler, query string) (int, string) {
	t.Helper()

	target := "/"
	if query != "" {
		target = "/?q=" + url.QueryEscape(query)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return rec.Code, rec.Body.String()
}

func search200(t *testing.T, query string) string {
	t.Helper()

	code, body := get(t, New(search.NewIndex(testDocs())), query)
	require.Equal(t, http.StatusOK, code)

	return body
}

func TestSearchPage(t *testing.T) {
	t.Parallel()

	t.Run("no query shows only the form", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "")

		assert.Contains(t, body, `name="q"`)
		assert.NotContains(t, body, `class="result"`)
		assert.NotContains(t, body, "No results")
	})

	t.Run("blank query counts as empty", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "   ")

		assert.NotContains(t, body, "No results")
	})

	t.Run("results in BM25 order", func(t *testing.T) {
		t.Parallel()

		idx := search.NewIndex(testDocs())
		results := idx.SearchBM25("file")
		require.Greater(t, len(results), 1)

		_, body := get(t, New(idx), "file")

		previous := -1
		for _, r := range results {
			position := strings.Index(body, `data-name="`+idx.Docs[r.DocID].Name+`"`)
			require.Greater(t, position, previous, idx.Docs[r.DocID].Name)
			previous = position
		}
	})

	t.Run("no results", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "missing")

		assert.Contains(t, body, "No results")
		assert.NotContains(t, body, `class="result"`)
	})

	t.Run("shows declaration and package", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "client")

		assert.Contains(t, tagPattern.ReplaceAllString(body, ""), "func (c *Client) Do(req *Request) (*Response, error)")
		assert.Contains(t, body, `<span class="tok-kw">func</span>`)
		assert.Contains(t, body, ">Client.Do<")
		assert.Contains(t, body, ">net/http<")
	})

	t.Run("only top result is open", func(t *testing.T) {
		t.Parallel()

		idx := search.NewIndex(testDocs())
		results := idx.SearchBM25("file")
		require.Greater(t, len(results), 1)

		_, body := get(t, New(idx), "file")

		assert.Contains(t, body, `data-name="`+idx.Docs[results[0].DocID].Name+`" open>`)
		assert.Equal(t, 1, strings.Count(body, `" open>`))
	})

	t.Run("highlights matched words", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "descriptor")

		assert.Contains(t, body, "<mark>descriptor</mark>")
	})

	t.Run("renders documentation markup", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "sends")

		assert.Contains(t, body, `client.<span class="tok-fn">Do</span>(req)`)
		assert.Contains(t, body, `href="https://pkg.go.dev/net/http#Client"`)
		assert.Contains(t, body, `href="https://pkg.go.dev/io#Reader"`)
	})

	t.Run("links to pkg.go.dev", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "sends")

		assert.Contains(t, body, `href="https://pkg.go.dev/net/http#Client.Do"`)
	})

	t.Run("no pkg.go.dev link for internal package", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "hidden")

		assert.NotContains(t, body, "pkg.go.dev/a/internal/b")
	})

	t.Run("caps results at 20 and shows total", func(t *testing.T) {
		t.Parallel()

		docs := make([]search.Document, 25)
		for i := range docs {
			name := "p.F" + strconv.Itoa(i)
			docs[i] = search.Document{Name: name, Package: "p", Text: "common word"}
		}

		_, body := get(t, New(search.NewIndex(docs)), "common")

		assert.Equal(t, 20, strings.Count(body, `class="result"`))
		assert.Contains(t, body, "20 of 25")
	})
}

func TestSearchPageEscaping(t *testing.T) {
	t.Parallel()

	t.Run("documentation text", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "evil")

		assert.NotContains(t, body, "<script>alert")
		assert.Contains(t, body, "&lt;script&gt;")
	})

	t.Run("declaration", func(t *testing.T) {
		t.Parallel()

		body := search200(t, "evil")

		assert.Contains(t, tagPattern.ReplaceAllString(body, ""), "c &lt;-chan int")
	})

	t.Run("query", func(t *testing.T) {
		t.Parallel()

		body := search200(t, `"><script>alert(2)</script>`)

		assert.NotContains(t, body, "<script>alert(2)")
	})
}

func TestRouting(t *testing.T) {
	t.Parallel()

	handler := New(search.NewIndex(testDocs()))

	tests := []struct {
		name   string
		method string
		target string
		want   int
	}{
		{"home", http.MethodGet, "/", http.StatusOK},
		{"unknown path", http.MethodGet, "/missing", http.StatusNotFound},
		{"POST not allowed", http.MethodPost, "/", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.target, nil))

			assert.Equal(t, tt.want, rec.Code)
		})
	}

	t.Run("content type", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	})
}

func TestPkgGoDevURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pkg    string
		symbol string
		want   string
	}{
		{"package overview", "net/http", "", "https://pkg.go.dev/net/http"},
		{"function", "os", "ReadFile", "https://pkg.go.dev/os#ReadFile"},
		{"method", "net/http", "Client.Do", "https://pkg.go.dev/net/http#Client.Do"},
		{"constant group", "x", "A, B", "https://pkg.go.dev/x#A"},
		{"dot in module path", "gopkg.in/yaml.v3", "Marshal", "https://pkg.go.dev/gopkg.in/yaml.v3#Marshal"},
		{"internal package", "a/internal/b", "F", ""},
		{"internal at end of path", "a/internal", "F", ""},
		{"internal as part of a name", "a/internals", "F", "https://pkg.go.dev/a/internals#F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, pkgGoDevURL(tt.pkg, tt.symbol))
		})
	}
}

func TestBestSentence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		query string
		want  string
	}{
		{"sentence with most matches", "First about files. Close idle connections now. Other.", "close idle", "Close idle connections now."},
		{"first sentence without matches", "First sentence. Second sentence.", "missing", "First sentence."},
		{"earlier sentence on a tie", "Read one. Read two.", "read", "Read one."},
		{"repeated word counts once", "Read read read. Read file.", "read file", "Read file."},
		{"collapses whitespace", "Close the\n\tconnection now. Other.", "connection", "Close the connection now."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, bestSentence(tt.text, tokenSet(tt.query)))
		})
	}

	t.Run("truncates long sentence by runes", func(t *testing.T) {
		t.Parallel()

		got := bestSentence(strings.Repeat("я", maxSnippetRunes+50), tokenSet("x"))

		assert.Len(t, []rune(got), maxSnippetRunes+1)
		assert.True(t, strings.HasSuffix(got, "…"))
	})
}

func TestHighlight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		query string
		want  string
	}{
		{"exact word", "read the file", "file", "read the <mark>file</mark>"},
		{"case insensitive", "Read it", "read", "<mark>Read</mark> it"},
		{"whole identifier", "ReadFile reads", "file", "<mark>ReadFile</mark> reads"},
		{"no partial word match", "profile", "file", "profile"},
		{"escapes match and text", "<b> & read", "read", "&lt;b&gt; &amp; <mark>read</mark>"},
		{"no matches", "<i>", "x", "&lt;i&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, string(highlight(tt.text, tokenSet(tt.query))))
		})
	}
}
