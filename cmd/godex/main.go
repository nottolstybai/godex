package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nottolstybai/godex/internal/extract"
	"github.com/nottolstybai/godex/internal/search"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: godex search [-pkg pattern] [-n limit] query")
	}
	if args[0] != "search" {
		return fmt.Errorf("unknown command %q", args[0])
	}

	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pkg := fs.String("pkg", "std", "какие пакеты искать")
	limit := fs.Int("n", 10, "сколько результатов показать")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return errors.New("query is required")
	}
	if *limit < 1 {
		return errors.New("-n must be positive")
	}

	docs, err := extract.Load(*pkg)
	if err != nil {
		return err
	}

	idx := search.NewIndex(docs)
	results := idx.SearchBM25(strings.Join(fs.Args(), " "))
	if len(results) == 0 {
		_, err := fmt.Fprintln(stdout, "no results")
		return err
	}

	for _, result := range results[:min(*limit, len(results))] {
		if _, err := fmt.Fprintf(stdout, "%.3f  %s\n", result.Score, idx.Docs[result.DocID].Name); err != nil {
			return err
		}
	}

	return nil
}
