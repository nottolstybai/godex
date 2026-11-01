# godex

Full-text search over Go documentation. Built to learn how TF-IDF and BM25 work

```bash
go run ./cmd/godex search close idle connections
go run ./cmd/godex serve
```

Use `-p` to choose packages, such as `-p ./...`. The default is the standard library.
