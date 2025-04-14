package source

import "io"

// NewReaderSource creates a new reader-based Source implementation that
// obtains content from the provided io.Reader.
func NewReaderSource(reader io.Reader) *ReaderSource {
	return &ReaderSource{reader: reader}
}

// ReaderSource implements Source for io.Reader content sources.
// It enables parsing enum definitions from any input that implements
// the io.Reader interface, such as network connections, string buffers,
// or custom data streams.
type ReaderSource struct {
	reader io.Reader
}

// Content reads the entire content from the underlying reader
// and returns it as a byte slice. This method consumes the reader,
// so the reader cannot be read from again.
func (rs *ReaderSource) Content() ([]byte, error) {
	return io.ReadAll(rs.reader)
}

// Filename returns a generic identifier for this source.
// Since reader sources typically don't have associated filenames,
// this returns the constant string "reader" to identify the source type.
func (rs *ReaderSource) Filename() string {
	return "reader"
}
