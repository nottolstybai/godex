package search

import (
	"cmp"
	"slices"
)

type Result struct {
	DocID int
	Score float64
}

func rankResults(scores map[int]float64) []Result {
	results := make([]Result, 0, len(scores))
	for docID, score := range scores {
		results = append(results, Result{DocID: docID, Score: score})
	}

	slices.SortFunc(results, func(a, b Result) int {
		return cmp.Or(
			cmp.Compare(b.Score, a.Score),
			cmp.Compare(a.DocID, b.DocID),
		)
	})

	return results
}
