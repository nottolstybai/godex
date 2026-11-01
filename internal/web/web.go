// Package web serves the search page.
package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/doc/comment"
	"go/token"
	"html"
	"html/template"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/nottolstybai/godex/internal/search"
)

const (
	maxResults      = 20
	maxSnippetRunes = 240
	pkgGoDevBaseURL = "https://pkg.go.dev"
)

//go:embed page.html
var pageHTML string

var page = template.Must(template.New("page").Parse(pageHTML))

var wordPattern = regexp.MustCompile(`[\p{L}\p{N}]+`)

type pageData struct {
	Query     string
	DocCount  int
	Total     int
	Results   []resultView
	HasSearch bool
}

type resultView struct {
	Name    string
	Title   string
	Package string
	Decl    template.HTML
	Snippet template.HTML
	Doc     template.HTML
	Link    string
	Score   string
}

// New returns a handler for the search page.
func New(idx *search.Index) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		if err := page.Execute(&body, newPageData(idx, r.URL.Query().Get("q"))); err != nil {
			http.Error(w, "render page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = body.WriteTo(w)
	})

	return mux
}

func newPageData(idx *search.Index, query string) pageData {
	query = strings.TrimSpace(query)

	data := pageData{Query: query, DocCount: len(idx.Docs), HasSearch: query != ""}
	if !data.HasSearch {
		return data
	}

	results := idx.SearchBM25(query)
	tokens := tokenSet(query)

	data.Total = len(results)
	data.Results = make([]resultView, 0, min(maxResults, len(results)))
	for _, r := range results[:min(maxResults, len(results))] {
		data.Results = append(data.Results, newResultView(idx.Docs[r.DocID], r.Score, tokens))
	}

	return data
}

func newResultView(d search.Document, score float64, tokens map[string]bool) resultView {
	symbol := strings.TrimPrefix(strings.TrimPrefix(d.Name, d.Package), ".")

	title := symbol
	if title == "" {
		title = d.Package
	}

	return resultView{
		Name:    d.Name,
		Title:   title,
		Package: d.Package,
		Decl:    highlightGo(d.Decl),
		Snippet: highlight(bestSentence(d.Text, tokens), tokens),
		Doc:     renderDoc(d.Text, d.Package),
		Link:    pkgGoDevURL(d.Package, symbol),
		Score:   fmt.Sprintf("%.2f", score),
	}
}

func tokenSet(query string) map[string]bool {
	tokens := make(map[string]bool)
	for _, token := range search.Tokenize(query) {
		tokens[token] = true
	}

	return tokens
}

// bestSentence returns the sentence matching the most distinct query tokens.
func bestSentence(text string, tokens map[string]bool) string {
	text = strings.Join(strings.Fields(text), " ")

	best, bestScore := "", -1
	for sentence := range strings.SplitAfterSeq(text, ". ") {
		sentence = strings.TrimSpace(sentence)

		matched := make(map[string]bool)
		for _, token := range search.Tokenize(sentence) {
			if tokens[token] {
				matched[token] = true
			}
		}

		if len(matched) > bestScore {
			best, bestScore = sentence, len(matched)
		}
	}

	if runes := []rune(best); len(runes) > maxSnippetRunes {
		return string(runes[:maxSnippetRunes]) + "…"
	}

	return best
}

// highlight escapes text and wraps matching words in <mark>.
func highlight(text string, tokens map[string]bool) template.HTML {
	var out strings.Builder

	last := 0
	for _, loc := range wordPattern.FindAllStringIndex(text, -1) {
		word := text[loc[0]:loc[1]]

		out.WriteString(html.EscapeString(text[last:loc[0]]))
		if slices.ContainsFunc(search.Tokenize(word), func(token string) bool { return tokens[token] }) {
			out.WriteString("<mark>")
			out.WriteString(html.EscapeString(word))
			out.WriteString("</mark>")
		} else {
			out.WriteString(html.EscapeString(word))
		}
		last = loc[1]
	}
	out.WriteString(html.EscapeString(text[last:]))

	return template.HTML(out.String()) //nolint:gosec // escaped above
}

// renderDoc renders a doc comment as HTML.
func renderDoc(text, pkg string) template.HTML {
	// Without LookupSym, [Client] stays plain text. Assume exported names exist.
	parser := comment.Parser{
		LookupSym: func(_, name string) bool { return token.IsExported(name) },
	}

	printer := comment.Printer{
		HeadingID: func(*comment.Heading) string { return "" },
		DocLinkURL: func(link *comment.DocLink) string {
			target := *link
			if target.ImportPath == "" {
				target.ImportPath = pkg
			}

			return target.DefaultURL(pkgGoDevBaseURL)
		},
	}

	var out strings.Builder
	for _, block := range parser.Parse(text).Content {
		if code, ok := block.(*comment.Code); ok {
			out.WriteString("<pre><code>")
			out.WriteString(string(highlightGo(strings.TrimRight(code.Text, "\n"))))
			out.WriteString("</code></pre>")
			continue
		}
		out.Write(printer.HTML(&comment.Doc{Content: []comment.Block{block}}))
	}

	return template.HTML(out.String()) //nolint:gosec // both sources escape text
}

// pkgGoDevURL returns the symbol's page, or "" for internal packages.
func pkgGoDevURL(pkg, symbol string) string {
	if slices.Contains(strings.Split(pkg, "/"), "internal") {
		return ""
	}

	link := pkgGoDevBaseURL + "/" + pkg
	if symbol == "" {
		return link
	}

	// A constant group links to its first name.
	anchor, _, _ := strings.Cut(symbol, ", ")

	return link + "#" + anchor
}
