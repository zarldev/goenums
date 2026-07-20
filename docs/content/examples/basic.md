---
title: Basic Examples
---

Learn how to use the core functionality of goenums with simple examples.

## Simple Enum Definition

Here's a simple enum definition for task statuses:

```go
package validation

type status int

//go:generate goenums status.go
const (
    unknown   status = iota // invalid Unknown
    pending                 // Pending
    approved                // Approved
    rejected                // Rejected
    completed               // Completed
)
```

This enum has a standard alias for each enum value defined in the comment.

Generate the enum implementations with the `goenums` tool:

```bash
$ go generate ./...
```

## Custom String Representations

As with the above example, we can add a custom string representation to each enum value:

### Standard Name Comment

```go
type ticketStatus int

//go:generate goenums status.go
const (
    unknown   ticketStatus = iota // invalid Unknown
    pending                       // Pending
    approved                      // Approved
    rejected                      // Rejected
    completed                     // Completed
)
```

### Name Comment with spaces
When using alias names that contain spaces, the double quotes are required:
```go
package validation

type ticketStatus int

//go:generate goenums status.go
const (
    unknown   ticketStatus = iota // invalid "Not Found"
    pending                       // "In Progress"
    approved                      // "Fully Approved"
    rejected                      // "Has Been Rejected"
    completed                     // "Successfully Completed"
)
```

## Case Insensitive String Parsing

Use the `-i` flag to enable case insensitive string parsing:

```go
//go:generate goenums -i status.go
```

Generated code will parse case insensitive strings. All of the below will validate and produce the `pending` enum:

```go
status, err := validation.ParseTicketStatus("In Progress")
if err != nil {
    fmt.Println("error:", err)
}
status, err := validation.ParseTicketStatus("in progress")
if err != nil {
    fmt.Println("error:", err)
}
status, err := validation.ParseTicketStatus("IN PROGRESS")
if err != nil {
    fmt.Println("error:", err)
}
```

## JSON, Text, Binary, YAML, and Database Storage

The generated enum type implements the:

* `json.Marshaler` and `json.Unmarshaler` interfaces
* `sql.Scanner` and `sql.Valuer` interfaces
* `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces
* `encoding.TextMarshaler` and `encoding.TextUnmarshaler` interfaces
* `yaml.Marshaler` and `yaml.Unmarshaler` interfaces

These interfaces allow you to use the enum type in JSON, text, binary, YAML, and database storage seamlessly.

## Selecting generated interfaces

By default, goenums generates JSON, Text, YAML, SQL, and Binary interface methods. Use `--interfaces` when you only want a subset:

```bash
goenums --interfaces=json,sql ticket_status.go
```

Valid values are `json`, `text`, `yaml`, `sql`, and `binary`. If `text` is not selected, `-legacy-text` has no generated-code effect.

## Enum accessor style

Generated enum values are exposed through a container such as `TicketStatuses`. The default `upper` style preserves the existing selector names:

```go
TicketStatuses.PENDING_APPROVAL
```

For idiomatic exported Go selectors, opt in when generating:

```go
//go:generate goenums --accessor-style=go ticket_status.go
```

```go
TicketStatuses.PendingApproval
```

Changing styles requires updating Go call sites. It does not change aliases, parsing, serialized values, database values, numeric values, or generated filenames.

## Text and JSON compatibility

`MarshalText` returns raw enum text by default. Use `goenums -legacy-text file.go` if you need the older quoted `MarshalText` output. JSON marshaling uses standard JSON string escaping, and JSON unmarshaling also accepts raw enum text when called directly.

## Numeric Parsing Support

The generated enums support parsing from various numeric types. Numeric inputs are matched against the underlying enum constant values, not ordinal positions:

```go
// Parse from different numeric types
status1, _ := validation.ParseTicketStatus(1)        // int
status2, _ := validation.ParseTicketStatus(int32(2)) // int32
status3, _ := validation.ParseTicketStatus(3.0)      // float64
status4, _ := validation.ParseTicketStatus(uint8(4)) // uint8

// All numeric types are supported: int, int8, int16, int32, int64,
// uint, uint8, uint16, uint32, uint64, float32, float64
```

# Basic Usage After Generation

Use the generated code in your Go project:

```go

ticketStatus := validation.TicketStatuses.PENDING

// Convert to string
fmt.Println(ticketStatus.String()) // "PENDING"

// Parse from various sources
input := "SKIPPED"
parsed, _ := validation.ParseTicketStatus(input)

// Validate enum values
if !parsed.IsValid() {
    fmt.Println("Invalid status")
}

// JSON marshaling/unmarshaling
type Task struct {
    ID     int              `json:"id"`
    Status validation.TicketStatus `json:"status"`
}

// Iterate through all values
for status := range validation.Statuses.All() {
    fmt.Printf("Status: %s\n", status)
}
```

[Back to Examples]({{< relref "/examples" >}})