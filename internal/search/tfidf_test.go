package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestIndex() *Index {
	return NewIndex([]Document{
		{Name: "d0", Text: "read all data from reader"},
		{Name: "d1", Text: "write data to file"},
		{Name: "d2", Text: "read file and return data"},
	})
}

// assertResults compares order exactly and scores within a tolerance.
func assertResults(t *testing.T, want, got []Result) {
	t.Helper()

	require.Len(t, got, len(want))
	for i := range want {
		assert.Equal(t, want[i].DocID, got[i].DocID, "position %d", i)
		assert.InDelta(t, want[i].Score, got[i].Score, 0.001, "position %d", i)
	}
}

func TestSearchTFIDF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  []Result
	}{
		{
			name:  "more matched words rank higher",
			query: "read file",
			want:  []Result{{DocID: 2, Score: 0.811}, {DocID: 0, Score: 0.405}, {DocID: 1, Score: 0.405}},
		},
		{
			name:  "single word",
			query: "write",
			want:  []Result{{DocID: 1, Score: 1.099}},
		},
		{
			name:  "one word not found",
			query: "write missing",
			want:  []Result{{DocID: 1, Score: 1.099}},
		},
		{
			name:  "word in every document scores zero",
			query: "data",
			want:  []Result{{DocID: 0, Score: 0}, {DocID: 1, Score: 0}, {DocID: 2, Score: 0}},
		},
		{
			name:  "ties break by lower DocID",
			query: "read data",
			want:  []Result{{DocID: 0, Score: 0.405}, {DocID: 2, Score: 0.405}, {DocID: 1, Score: 0}},
		},
		{
			name:  "repeated query word counts twice",
			query: "read read",
			want:  []Result{{DocID: 0, Score: 0.811}, {DocID: 2, Score: 0.811}},
		},
		{
			name:  "query is tokenized",
			query: "READ, File!",
			want:  []Result{{DocID: 2, Score: 0.811}, {DocID: 0, Score: 0.405}, {DocID: 1, Score: 0.405}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertResults(t, tt.want, newTestIndex().SearchTFIDF(tt.query))
		})
	}

	t.Run("term frequency multiplies score", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]Document{
			{Text: "read read file"},
			{Text: "write file"},
		})

		want := []Result{{DocID: 0, Score: 1.386}}
		assertResults(t, want, idx.SearchTFIDF("read"))
	})
}

func TestSearchTFIDFEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
	}{
		{"word not in index", "missing"},
		{"empty query", ""},
		{"separators only", " ,.!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, newTestIndex().SearchTFIDF(tt.query))
		})
	}

	t.Run("empty index", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, NewIndex(nil).SearchTFIDF("read"))
	})
}
