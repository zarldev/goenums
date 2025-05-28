---
layout: default
title: goenums
---

`goenums` addresses Go's lack of native enum support by generating comprehensive, type-safe enum implementations from simple constant declarations. Transform basic `iota` based constants into feature-rich enums with string conversion, validation, JSON handling, database integration, and more.

## Key Features

- **Type Safety**: Wrapper types prevent accidental misuse of enum values
- **String Conversion**: Automatic string representation and parsing
- **JSON Support**: Built-in marshaling and unmarshaling 
- **Database Integration**: SQL Scanner and Valuer implementations
- **Validation**: Methods to check for valid enum values
- **Iteration**: Modern Go 1.23+ iteration support with legacy fallback
- **Extensibility**: Add custom fields to enums via comments
- **Exhaustive Handling**: Helper functions to ensure you handle all enum values
- **Zero Dependencies**: Completely dependency-free, using only the Go standard library

Get Started with [Installation]({{ '/installation' | relative_url }}) → 