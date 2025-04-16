// The gofile package provides Go-specific parsing and generation capabilities.
//
// This package contains two primary components:
//
//   - Parser: Creates enum representations from Go constant declarations
//   - Generator: Produces Go code for type-safe enum implementations
//
// # Parser
//
// The Parser analyzes Go constant declarations and extracts enum information.
// It recognizes:
//   - iota-based constant groups
//   - Comment-based metadata
//   - Enum value properties
//   - Invalid enum markers
//
// # Generator
//
// The Generator produces Go code for enum implementations with:
//   - Type-safe wrapper types
//   - String conversion and parsing
//   - JSON and database integration
//   - Validation methods
//   - Iteration utilities (with legacy support)
//
// Both components operate on abstract interfaces, maintaining the system's
// extensible architecture.
package gofile
