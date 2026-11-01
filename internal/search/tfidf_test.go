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

// assertResults сравнивает номера документов и их порядок точно, а очки — с допуском.
func assertResults(t *testing.T, want, got []Result) {
	t.Helper()

	require.Len(t, got, len(want))
	for i := range want {
		assert.Equal(t, want[i].DocID, got[i].DocID, "позиция %d", i)
		assert.InDelta(t, want[i].Score, got[i].Score, 0.001, "позиция %d", i)
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
			name:  "больше совпавших слов — выше в выдаче",
			query: "read file",
			want:  []Result{{DocID: 2, Score: 0.811}, {DocID: 0, Score: 0.405}, {DocID: 1, Score: 0.405}},
		},
		{
			name:  "одно слово",
			query: "write",
			want:  []Result{{DocID: 1, Score: 1.099}},
		},
		{
			name:  "одно из слов не найдено",
			query: "write missing",
			want:  []Result{{DocID: 1, Score: 1.099}},
		},
		{
			name:  "слово из всех документов получает ноль",
			query: "data",
			want:  []Result{{DocID: 0, Score: 0}, {DocID: 1, Score: 0}, {DocID: 2, Score: 0}},
		},
		{
			name:  "при равных очках раньше меньший DocID",
			query: "read data",
			want:  []Result{{DocID: 0, Score: 0.405}, {DocID: 2, Score: 0.405}, {DocID: 1, Score: 0}},
		},
		{
			name:  "повтор слова в запросе считается дважды",
			query: "read read",
			want:  []Result{{DocID: 0, Score: 0.811}, {DocID: 2, Score: 0.811}},
		},
		{
			name:  "запрос проходит через токенизатор",
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

	t.Run("частота слова в документе умножает очки", func(t *testing.T) {
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
		{"слова нет в индексе", "missing"},
		{"пустой запрос", ""},
		{"только разделители", " ,.!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, newTestIndex().SearchTFIDF(tt.query))
		})
	}

	t.Run("пустой индекс", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, NewIndex(nil).SearchTFIDF("read"))
	})
}
