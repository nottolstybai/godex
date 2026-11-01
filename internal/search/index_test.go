package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndex(t *testing.T) {
	t.Parallel()

	t.Run("три документа", func(t *testing.T) {
		t.Parallel()

		docs := []Document{
			{Name: "d0", Text: "read all data from reader"},
			{Name: "d1", Text: "write data to file"},
			{Name: "d2", Text: "read file and return data"},
		}

		idx := NewIndex(docs)

		assert.Equal(t, []Posting{{DocID: 0, TF: 1}, {DocID: 2, TF: 1}}, idx.Postings["read"])
		assert.Equal(t, []Posting{{DocID: 1, TF: 1}, {DocID: 2, TF: 1}}, idx.Postings["file"])
		assert.Equal(t, []Posting{{DocID: 0, TF: 1}, {DocID: 1, TF: 1}, {DocID: 2, TF: 1}}, idx.Postings["data"])
		assert.NotContains(t, idx.Postings, "missing")
		assert.Equal(t, []int{5, 4, 5}, idx.DocLen)
		assert.InDelta(t, 14.0/3.0, idx.AvgLen, 0.001)
		assert.Equal(t, docs, idx.Docs)
	})

	t.Run("повторы слова в документе", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]Document{{Text: "read read file"}})

		assert.Equal(t, []Posting{{DocID: 0, TF: 2}}, idx.Postings["read"])
		assert.Equal(t, []int{3}, idx.DocLen)
		assert.InDelta(t, 3.0, idx.AvgLen, 0.001)
	})

	t.Run("нет документов", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex(nil)

		assert.Zero(t, idx.AvgLen)
		assert.Empty(t, idx.Postings)
		assert.Empty(t, idx.DocLen)
	})
}
