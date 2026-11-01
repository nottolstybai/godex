package search

import "math"

func (idx *Index) SearchTFIDF(query string) []Result {
	scores := make(map[int]float64)
	for _, token := range Tokenize(query) {
		postings := idx.Postings[token]
		if len(postings) == 0 {
			continue
		}

		idf := math.Log(float64(len(idx.Docs)) / float64(len(postings)))

		for _, posting := range postings {
			scores[posting.DocID] += float64(posting.TF) * idf
		}
	}

	return rankResults(scores)
}
