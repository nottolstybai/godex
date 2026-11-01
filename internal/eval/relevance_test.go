// Package eval measures search quality on standard library docs.
// Run with -v to see the rank of each query.
package eval

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nottolstybai/godex/internal/extract"
	"github.com/nottolstybai/godex/internal/search"
)

// Any document in want counts as a hit.
var queries = []struct {
	text string
	want []string
}{
	{"split string by whitespace", []string{"strings.Fields", "strings.FieldsSeq"}},
	{"close idle connections", []string{"net/http.Transport.CloseIdleConnections", "net/http.Client.CloseIdleConnections"}},
	{"read entire file into memory", []string{"os.ReadFile"}},
	{"read all data from a reader", []string{"io.ReadAll"}},
	{"sort a slice with a custom comparison function", []string{"slices.SortFunc", "slices.SortStableFunc", "sort.Slice"}},
	{"parse a time string", []string{"time.Parse", "time.ParseInLocation"}},
	{"convert string to integer", []string{"strconv.Atoi", "strconv.ParseInt"}},
	{"get the current time", []string{"time.Now"}},
	{"make an http get request", []string{"net/http.Get", "net/http.Client.Get"}},
	{"encode a value as json", []string{"encoding/json.Marshal", "encoding/json.Encoder.Encode"}},
	{"decode json into a struct", []string{"encoding/json.Unmarshal", "encoding/json.Decoder.Decode"}},
	{"cancel a context after a timeout", []string{"context.WithTimeout", "context.WithDeadline"}},
	{"wait for goroutines to finish", []string{"sync.WaitGroup", "sync.WaitGroup.Wait"}},
	{"mutual exclusion lock", []string{"sync.Mutex"}},
	{"run a function only once", []string{"sync.Once", "sync.Once.Do", "sync.OnceFunc"}},
	{"walk a directory tree", []string{"path/filepath.WalkDir", "path/filepath.Walk", "io/fs.WalkDir"}},
	{"read an environment variable", []string{"os.Getenv", "os.LookupEnv"}},
	{"create a temporary file", []string{"os.CreateTemp"}},
	{"join path elements", []string{"path/filepath.Join", "path.Join"}},
	{"replace all occurrences in a string", []string{"strings.ReplaceAll", "strings.Replace"}},
	{"compile a regular expression", []string{"regexp.Compile", "regexp.MustCompile"}},
	{"generate a random number", []string{"math/rand/v2.IntN", "math/rand.Intn", "math/rand/v2.Int", "math/rand.Int", "crypto/rand.Int"}},
	{"compute sha256 checksum", []string{"crypto/sha256.Sum256"}},
	{"base64 encoding", []string{"encoding/base64", "encoding/base64.Encoding.EncodeToString"}},
	{"listen for tcp connections", []string{"net.Listen", "net.ListenTCP"}},
	{"start an http server", []string{"net/http.ListenAndServe", "net/http.Server.ListenAndServe"}},
	{"execute an external command", []string{"os/exec.Command", "os/exec"}},
	{"reverse the order of elements", []string{"slices.Reverse"}},
	{"check whether a string has a prefix", []string{"strings.HasPrefix"}},
	{"remove duplicates from a sorted slice", []string{"slices.Compact"}},
	{"compress data with gzip", []string{"compress/gzip", "compress/gzip.NewWriter"}},
	{"wrap an error with a message", []string{"fmt.Errorf"}},
}

// TestRelevance guards against quality regressions. Thresholds sit below
// current values because docs change between Go versions.
func TestRelevance(t *testing.T) {
	t.Parallel()

	docs, err := extract.Load("std")
	require.NoError(t, err)

	idx := search.NewIndex(docs)

	var hitsTop1, hitsTop10 int
	var reciprocalRanks float64

	for _, q := range queries {
		results := idx.SearchBM25(q.text)

		// rank is 0 if no expected document is found.
		rank := slices.IndexFunc(results, func(r search.Result) bool {
			return slices.Contains(q.want, idx.Docs[r.DocID].Name)
		}) + 1

		t.Logf("%4d  %s", rank, q.text)

		if rank == 0 {
			continue
		}
		reciprocalRanks += 1 / float64(rank)
		if rank == 1 {
			hitsTop1++
		}
		if rank <= 10 {
			hitsTop10++
		}
	}

	mrr := reciprocalRanks / float64(len(queries))
	t.Logf("top 1: %d, top 10: %d of %d, MRR: %.3f", hitsTop1, hitsTop10, len(queries), mrr)

	assert.GreaterOrEqual(t, hitsTop1, 13)
	assert.GreaterOrEqual(t, hitsTop10, 23)
	assert.GreaterOrEqual(t, mrr, 0.55)
}
