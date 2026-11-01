package search

import "math"

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

func (idx *Index) SearchBM25(query string) []Result {
	scores := make(map[int]float64)
	for _, token := range Tokenize(query) {
		postings := idx.Postings[token]
		if len(postings) == 0 {
			continue
		}

		docNum := float64(len(idx.Docs))
		df := float64(len(postings))
		idf := math.Log(1 + (docNum-df+0.5)/(df+0.5))

		for _, posting := range postings {
			tf := float64(posting.TF)
			norm := 1 - bm25B + bm25B*float64(idx.DocLen[posting.DocID])/idx.AvgLen
			scores[posting.DocID] += idf * tf * (bm25K1 + 1) / (tf + bm25K1*norm)
		}
	}

	return rankResults(scores)
}
