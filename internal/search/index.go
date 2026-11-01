package search

type Document struct {
	Name    string // "net/http.Client.Do"
	Package string // "net/http"
	Decl    string // "func (c *Client) Do(req *Request) (*Response, error)"
	Text    string
}

type Posting struct {
	DocID int
	TF    int
}

type Index struct {
	Docs     []Document
	Postings map[string][]Posting
	DocLen   []int
	AvgLen   float64
}

func NewIndex(docs []Document) *Index {
	postings := make(map[string][]Posting)
	docLen := make([]int, 0, len(docs))
	avgLen := 0.0
	totalTokens := 0

	for id, doc := range docs {
		tokens := Tokenize(doc.Text)
		docLen = append(docLen, len(tokens))
		totalTokens += len(tokens)

		tokenOccurs := make(map[string]int, len(tokens))
		for _, token := range tokens {
			tokenOccurs[token] += 1
		}

		for token, tf := range tokenOccurs {
			postings[token] = append(postings[token], Posting{DocID: id, TF: tf})
		}
	}

	if len(docs) > 0 {
		avgLen = float64(totalTokens) / float64(len(docs))
	}

	return &Index{
		Docs:     docs,
		Postings: postings,
		DocLen:   docLen,
		AvgLen:   avgLen,
	}
}
