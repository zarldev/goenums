package source

import "os"

// NewFileSource creates a new file-based Source implementation that reads
// enum definitions from a file at the specified path.
func NewFileSource(path string) *FileSource {
	return &FileSource{Path: path}
}

// FileSource implements Source for file-based content sources.
// It reads enum definitions from a file on the local filesystem.
type FileSource struct {
	// Path is the filesystem path to the source file
	Path string
}

// Content reads and returns the file's contents as a byte slice.
// It fulfills the Source interface by providing the raw content
// to be parsed for enum definitions.
func (fs *FileSource) Content() ([]byte, error) {
	return os.ReadFile(fs.Path)
}

// Filename returns the path of the source file.
// This identifies the source in error messages and generated
// code documentation.
func (fs *FileSource) Filename() string {
	return fs.Path
}
