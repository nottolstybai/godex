package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchBM25(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  []Result
	}{
		{
			name:  "shorter document ranks higher at equal TF",
			query: "read file",
			want:  []Result{{DocID: 2, Score: 0.913}, {DocID: 1, Score: 0.499}, {DocID: 0, Score: 0.457}},
		},
		{
			name:  "single word",
			query: "write",
			want:  []Result{{DocID: 1, Score: 1.042}},
		},
		{
			name:  "one word not found",
			query: "write missing",
			want:  []Result{{DocID: 1, Score: 1.042}},
		},
		{
			name:  "word in every document scores above zero",
			query: "data",
			want:  []Result{{DocID: 1, Score: 0.142}, {DocID: 0, Score: 0.13}, {DocID: 2, Score: 0.13}},
		},
		{
			name:  "scores add up across words",
			query: "file data",
			want:  []Result{{DocID: 1, Score: 0.641}, {DocID: 2, Score: 0.586}, {DocID: 0, Score: 0.13}},
		},
		{
			name:  "repeated query word counts twice",
			query: "read read",
			want:  []Result{{DocID: 0, Score: 0.913}, {DocID: 2, Score: 0.913}},
		},
		{
			name:  "query is tokenized",
			query: "READ, File!",
			want:  []Result{{DocID: 2, Score: 0.913}, {DocID: 1, Score: 0.499}, {DocID: 0, Score: 0.457}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertResults(t, tt.want, newTestIndex().SearchBM25(tt.query))
		})
	}

	t.Run("ties break by lower DocID", func(t *testing.T) {
		t.Parallel()

		// Map order is random, so a single run can pass without the tiebreak.
		idx := newTestIndex()
		want := []Result{{DocID: 1, Score: 0.142}, {DocID: 0, Score: 0.13}, {DocID: 2, Score: 0.13}}
		for range 500 {
			assertResults(t, want, idx.SearchBM25("data"))
		}
	})

	t.Run("term frequency saturates", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]Document{
			{Text: "read read read read read read read read read read"},
			{Text: "read a b c d e f g h i"},
			{Text: "other a b c d e f g h i"},
		})

		want := []Result{{DocID: 0, Score: 0.923}, {DocID: 1, Score: 0.47}}
		assertResults(t, want, idx.SearchBM25("read"))
	})

	t.Run("match in a shorter document scores higher", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]Document{
			{Text: "read file"},
			{Text: "read a b c d e f g h i"},
			{Text: "other a b c d e f g h i"},
		})

		want := []Result{{DocID: 0, Score: 0.669}, {DocID: 1, Score: 0.409}}
		assertResults(t, want, idx.SearchBM25("read"))
	})
}

func TestSearchBM25Empty(t *testing.T) {
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

			assert.Empty(t, newTestIndex().SearchBM25(tt.query))
		})
	}

	t.Run("empty index", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, NewIndex(nil).SearchBM25("read"))
	})
}
