// The source package provides implementations of the enum.Source interface.
//
// These implementations abstract the origin of input content, allowing the system
// to process enums from various sources without changing the core logic. By
// separating content retrieval from content processing, it enables a more
// flexible and extensible architecture.
//
// Current implementations include:
//   - FileSource: Retrieves content from the provided filesystem
//   - ReaderSource: Obtains content from an io.Reader
//
// Using this abstraction, parsers can focus solely on the transformation of
// content into enum representations, without concern for the content's origin.
// This clean separation of responsibilities makes it easy to add support for
// new content sources without modifying parsing logic.
//
// Each Source implementation provides standardized Content() and Filename()
// methods, ensuring consistent behavior regardless of the underlying source.
package source
