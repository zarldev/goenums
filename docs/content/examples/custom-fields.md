---
title: Custom Fields
---

# Custom Fields

Add structured data to each enum value by declaring fields in the enum comments.

```go
package example

type priority int

const (
	low priority = iota // Low | Weight: 1
	medium              // Medium | Weight: 5
	high                // High | Weight: 10
)
```

Generate the enum with matching field definitions:

```bash
goenums priority.go
```

The generated wrapper retains the enum value and exposes the declared fields, allowing callers to use strongly typed metadata alongside normal parsing and serialization methods.

See the [usage guide]({{< relref "/usage" >}}) for field declaration details.
