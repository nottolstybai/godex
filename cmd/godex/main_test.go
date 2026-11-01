package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nottolstybai/godex/internal/extract"
	"github.com/nottolstybai/godex/internal/search"
)

const samplePkg = "../../internal/extract/testdata/sample"

func runOK(t *testing.T, args ...string) string {
	t.Helper()

	var out bytes.Buffer
	require.NoError(t, run(args, &out))

	return out.String()
}

// bm25Output строит ожидаемую выдачу из уже проверенных extract и search.
func bm25Output(t *testing.T, pattern, query string, limit int) string {
	t.Helper()

	docs, err := extract.Load(pattern)
	require.NoError(t, err)

	idx := search.NewIndex(docs)
	results := idx.SearchBM25(query)

	var out strings.Builder
	for _, r := range results[:min(limit, len(results))] {
		fmt.Fprintf(&out, "%.3f  %s\n", r.Score, idx.Docs[r.DocID].Name)
	}

	return out.String()
}

func lines(output string) []string {
	return strings.Split(strings.TrimSpace(output), "\n")
}

func TestRunSearch(t *testing.T) {
	t.Parallel()

	t.Run("формат строки выдачи", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "whitespace")

		assert.Regexp(t, `^\d+\.\d{3}  github\.com/nottolstybai/godex/internal/extract/testdata/sample\.Split\n$`, got)
	})

	t.Run("выдача BM25 по убыванию очков", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "reader")

		assert.Equal(t, bm25Output(t, samplePkg, "reader", 10), got)
	})

	t.Run("слова запроса без кавычек склеиваются", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "split", "reader")

		assert.Equal(t, bm25Output(t, samplePkg, "split reader", 10), got)
	})

	t.Run("запрос в кавычках", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "split reader")

		assert.Equal(t, bm25Output(t, samplePkg, "split reader", 10), got)
	})

	t.Run("флаг -n ограничивает выдачу", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "-n", "2", "reader")

		assert.Equal(t, bm25Output(t, samplePkg, "reader", 2), got)
		assert.Len(t, lines(got), 2)
	})

	t.Run("-n больше числа результатов", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "-n", "100", "whitespace")

		assert.Len(t, lines(got), 1)
	})

	t.Run("по умолчанию 10 результатов", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", "net/http", "the")

		assert.Len(t, lines(got), 10)
	})

	t.Run("по умолчанию ищет в стандартной библиотеке", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-n", "1000", "white space")

		assert.Contains(t, got, "  strings.Fields\n")
	})

	t.Run("ничего не найдено", func(t *testing.T) {
		t.Parallel()

		got := runOK(t, "search", "-pkg", samplePkg, "missing")

		assert.Equal(t, "no results\n", got)
	})
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"нет команды", nil},
		{"неизвестная команда", []string{"index", "./..."}},
		{"нет запроса", []string{"search", "-pkg", samplePkg}},
		{"неизвестный флаг", []string{"search", "-unknown", "reader"}},
		{"-n равен нулю", []string{"search", "-pkg", samplePkg, "-n", "0", "reader"}},
		{"-n отрицательный", []string{"search", "-pkg", samplePkg, "-n", "-1", "reader"}},
		{"пакета не существует", []string{"search", "-pkg", "./missing", "reader"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			assert.Error(t, run(tt.args, &out))
		})
	}
}
