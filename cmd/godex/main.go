package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/nottolstybai/godex/internal/extract"
	"github.com/nottolstybai/godex/internal/search"
	"github.com/nottolstybai/godex/internal/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := &cobra.Command{
		Use:               "godex",
		Short:             "Search inside Go documentation",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.AddCommand(newSearchCmd(), newServeCmd())

	return root.ExecuteContext(ctx)
}

func newSearchCmd() *cobra.Command {
	var (
		pkg   string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "search query...",
		Short: "Print the best matches for a query",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit < 1 {
				return errors.New("--limit must be positive")
			}

			docs, err := extract.Load(pkg)
			if err != nil {
				return err
			}

			idx := search.NewIndex(docs)
			results := idx.SearchBM25(strings.Join(args, " "))
			if len(results) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no results")
				return err
			}

			for _, result := range results[:min(limit, len(results))] {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%.3f  %s\n", result.Score, idx.Docs[result.DocID].Name); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&pkg, "pkg", "p", "std", "packages to search, as for go list")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "maximum number of results")

	return cmd
}

func newServeCmd() *cobra.Command {
	var pkg, addr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the search web page",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Bind first to fail fast on a bad address.
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			defer listener.Close()

			docs, err := extract.Load(pkg)
			if err != nil {
				return err
			}

			server := &http.Server{
				Handler:           web.New(search.NewIndex(docs)),
				ReadHeaderTimeout: 5 * time.Second,
			}

			shutdownDone := make(chan error, 1)
			go func() {
				<-cmd.Context().Done()

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				shutdownDone <- server.Shutdown(ctx)
			}()

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%d documents indexed, open http://%s\n", len(docs), listener.Addr()); err != nil {
				return err
			}

			if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
				return err
			}

			return <-shutdownDone
		},
	}

	cmd.Flags().StringVarP(&pkg, "pkg", "p", "std", "packages to index, as for go list")
	cmd.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:8080", "address to listen on")

	return cmd
}
