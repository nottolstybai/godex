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
			name:  "при равном TF короткий документ выше длинного",
			query: "read file",
			want:  []Result{{DocID: 2, Score: 0.913}, {DocID: 1, Score: 0.499}, {DocID: 0, Score: 0.457}},
		},
		{
			name:  "одно слово",
			query: "write",
			want:  []Result{{DocID: 1, Score: 1.042}},
		},
		{
			name:  "одно из слов не найдено",
			query: "write missing",
			want:  []Result{{DocID: 1, Score: 1.042}},
		},
		{
			name:  "слово из всех документов получает небольшой плюс, а не ноль",
			query: "data",
			want:  []Result{{DocID: 1, Score: 0.142}, {DocID: 0, Score: 0.13}, {DocID: 2, Score: 0.13}},
		},
		{
			name:  "очки за разные слова складываются",
			query: "file data",
			want:  []Result{{DocID: 1, Score: 0.641}, {DocID: 2, Score: 0.586}, {DocID: 0, Score: 0.13}},
		},
		{
			name:  "повтор слова в запросе считается дважды",
			query: "read read",
			want:  []Result{{DocID: 0, Score: 0.913}, {DocID: 2, Score: 0.913}},
		},
		{
			name:  "запрос проходит через токенизатор",
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

	t.Run("при равных очках раньше меньший DocID", func(t *testing.T) {
		t.Parallel()

		// Порядок обхода map случаен, поэтому один запуск может пройти и без DocID в сортировке.
		idx := newTestIndex()
		want := []Result{{DocID: 1, Score: 0.142}, {DocID: 0, Score: 0.13}, {DocID: 2, Score: 0.13}}
		for range 500 {
			assertResults(t, want, idx.SearchBM25("data"))
		}
	})

	t.Run("насыщение: 10 упоминаний не дают в 10 раз больше очков", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]Document{
			{Text: "read read read read read read read read read read"},
			{Text: "read a b c d e f g h i"},
			{Text: "other a b c d e f g h i"},
		})

		want := []Result{{DocID: 0, Score: 0.923}, {DocID: 1, Score: 0.47}}
		assertResults(t, want, idx.SearchBM25("read"))
	})

	t.Run("длина: совпадение в коротком документе весит больше", func(t *testing.T) {
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
		{"слова нет в индексе", "missing"},
		{"пустой запрос", ""},
		{"только разделители", " ,.!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, newTestIndex().SearchBM25(tt.query))
		})
	}

	t.Run("пустой индекс", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, NewIndex(nil).SearchBM25("read"))
	})
}
