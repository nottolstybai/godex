// Package sample is used to test documentation extraction.
package sample

// MaxSize is the largest allowed size.
const MaxSize = 10

// Buffer limits for a reader.
const (
	MinBuffer = 1
	MaxBuffer = 2
)

const NoDocConst = 1

// Timeout is the default timeout in seconds.
var Timeout = 30

// Split divides a string by whitespace.
func Split(s string) []string { return nil }

func NoDocFunc() {}

// hidden is unexported and must not be extracted.
func hidden() {}

// Reader reads data from a source.
type Reader struct{}

// DefaultReader is the shared reader.
var DefaultReader *Reader = NewReader()

// NewReader creates a Reader.
func NewReader() *Reader {
	hidden()
	return &Reader{}
}

// Close releases the reader.
func (r *Reader) Close() error { return nil }

func (r *Reader) NoDocMethod() {}

// Mode selects a speed.
type Mode int

// Fast is the quick mode.
const Fast Mode = 1

type Bare struct{}

// Run starts the work.
func (Bare) Run() {}
