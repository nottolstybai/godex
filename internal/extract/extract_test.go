package extract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nottolstybai/godex/internal/search"
)

const testdataImportPath = "github.com/nottolstybai/godex/internal/extract/testdata"

func sampleDocs(importPath string) []search.Document {
	doc := func(suffix, decl, text string) search.Document {
		return search.Document{Name: importPath + suffix, Package: importPath, Decl: decl, Text: text}
	}

	return []search.Document{
		doc("", "", "Package sample is used to test documentation extraction."),
		doc(".MaxSize", "const MaxSize = 10", "MaxSize is the largest allowed size."),
		doc(".MinBuffer, MaxBuffer", "const (\n\tMinBuffer = 1\n\tMaxBuffer = 2\n)", "Buffer limits for a reader."),
		doc(".Timeout", "var Timeout = 30", "Timeout is the default timeout in seconds."),
		doc(".Split", "func Split(s string) []string", "Split divides a string by whitespace."),
		doc(".Reader", "type Reader struct{}", "Reader reads data from a source."),
		doc(".DefaultReader", "var DefaultReader *Reader = NewReader()", "DefaultReader is the shared reader."),
		doc(".NewReader", "func NewReader() *Reader", "NewReader creates a Reader."),
		doc(".Reader.Close", "func (r *Reader) Close() error", "Close releases the reader."),
		doc(".Mode", "type Mode int", "Mode selects a speed."),
		doc(".Fast", "const Fast Mode = 1", "Fast is the quick mode."),
		doc(".Bare.Run", "func (Bare) Run()", "Run starts the work."),
	}
}

func TestParsePackage(t *testing.T) {
	t.Parallel()

	t.Run("extracts all documented symbols", func(t *testing.T) {
		t.Parallel()

		got, err := parsePackage("testdata/sample", "example.com/sample", []string{"sample.go"})

		require.NoError(t, err)
		assert.ElementsMatch(t, sampleDocs("example.com/sample"), got)
	})

	t.Run("no files", func(t *testing.T) {
		t.Parallel()

		got, err := parsePackage("testdata/sample", "example.com/sample", nil)

		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		_, err := parsePackage("testdata/sample", "example.com/sample", []string{"missing.go"})

		assert.Error(t, err)
	})

	t.Run("syntax error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.go"), []byte("package broken\n\nfunc Broken( {\n"), 0o644))

		_, err := parsePackage(dir, "example.com/broken", []string{"broken.go"})

		assert.Error(t, err)
	})
}

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("single package skips build-tagged file", func(t *testing.T) {
		t.Parallel()

		got, err := Load("./testdata/sample")

		require.NoError(t, err)
		assert.ElementsMatch(t, sampleDocs(testdataImportPath+"/sample"), got)
	})

	t.Run("multiple packages by pattern", func(t *testing.T) {
		t.Parallel()

		got, err := Load("net/http/...")
		require.NoError(t, err)

		names := make([]string, 0, len(got))
		for _, d := range got {
			names = append(names, d.Name)
		}

		assert.Contains(t, names, "net/http.NewRequest")
		assert.Contains(t, names, "net/http/httptest.NewServer")
	})

	t.Run("hides internal std packages", func(t *testing.T) {
		t.Parallel()

		for _, pattern := range []string{"net/http/internal", "vendor/golang.org/x/net/http2/hpack"} {
			got, err := Load(pattern)

			require.NoError(t, err)
			assert.Empty(t, got, pattern)
		}
	})

	t.Run("missing package", func(t *testing.T) {
		t.Parallel()

		_, err := Load("./testdata/missing")

		assert.Error(t, err)
	})

	t.Run("real package: constructor, method, overview", func(t *testing.T) {
		t.Parallel()

		got, err := Load("net/http")
		require.NoError(t, err)

		texts := make(map[string]string, len(got))
		for _, d := range got {
			texts[d.Name] = d.Text
		}

		assert.Contains(t, texts["net/http.NewRequest"], "NewRequest wraps")
		assert.Contains(t, texts["net/http.Client.Do"], "Do sends an HTTP request")
		assert.Contains(t, texts["net/http"], "Package http provides")
	})
}
