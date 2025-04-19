---
layout: default
title: Basic Examples
---

# Basic Examples

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

## JSON & Database Storage

The generated enum type also implements the `json.Unmarshal` and `json.Marshal` interfaces along with the `sql.Scanner` and `sql.Valuer` interfaces to handle parsing over the wire via HTTP or via a Database.

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

[Back to Examples]({{ '/examples' | relative_url }})