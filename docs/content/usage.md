---
title: Usage
---

Using goenums follows a simple workflow:

1. Define enum constants with Go's `iota`.
2. Add a `go:generate` directive.
3. Run `go generate`.
4. Use the generated wrapper and enum container in your code.

## Command-line options

```text
Usage: goenums [options] file.go[,file2.go,...]

Options:
  -accessor-style string
        Set enum accessor naming style: upper or go (default: upper)
  -constraints
        Generate local numeric constraints instead of importing golang.org/x/exp/constraints (default: true)
  -x-exp-constraints
        Import golang.org/x/exp/constraints instead of generating local numeric constraints (default: false)
  -failfast
        Enable failfast mode - fail on generation of invalid enum while parsing (default: false)
  -help
        Print help information
  -insensitive
        Generate case insensitive string parsing (default: false)
  -interfaces string
        Generate only the listed interface handlers: json,text,yaml,sql,binary (default: all)
  -legacy
        Generate legacy code without Go 1.23+ iterator support (default: false)
  -legacy-text
        Generate legacy quoted MarshalText output (default: false)
  -output string
        Specify the output format (default: go)
  -version
        Print version information
  -verbose
        Enable verbose mode - prints out the generated code (default: false)
```

For existing generated enums, see [Migrating to goenums v0.6]({{< relref "/migration-v0.6" >}}) for the `-legacy-text` compatibility option.
## Add a `go:generate` directive

Add a directive to the Go file that declares the enum:

```go
package validation

type status int

//go:generate goenums status.go
const (
	unknown status = iota // invalid Unknown
	pending               // Pending
	approved              // Approved
	rejected              // Rejected
	completed             // Completed
)
```

Generate all directives in your module:

```bash
go generate ./...
```

For a `status` type, goenums writes `statuses_enums.go` alongside the source file.

## Select generated interfaces

By default, goenums generates JSON, Text, YAML, SQL, and Binary interface methods. Select only the handlers your package needs:

```bash
goenums --interfaces=json,sql status.go
```

Valid values are `json`, `text`, `yaml`, `sql`, and `binary`. If the flag is omitted, all interfaces are generated. `-legacy-text` affects only Text handlers.

## Select enum accessor style

The default preserves the historical uppercase container selectors:

```go
Statuses.PENDING_APPROVAL
```

New projects can opt into idiomatic exported Go selectors:

```bash
goenums --accessor-style=go status.go
```

```go
Statuses.PendingApproval
```

Changing accessor style requires updating Go call sites. It does not change aliases, parsing, serialized values, database values, numeric values, or generated filenames.

## Use generated code

Generated code includes a wrapper type, a container of enum values, parsing and validation methods, iteration helpers, and the selected interface implementations.

```go
import "yourpackage/validation"

status := validation.Statuses.PASSED
fmt.Println(status.String())

parsed, err := validation.ParseStatus("FAILED")
if err != nil {
	return err
}

for status := range validation.Statuses.All() {
	fmt.Printf("Status: %s\n", status)
}
```

## Compile-time exhaustiveness

Each enum has a matcher interface and `Match`/`MustMatch` functions. Implement every matcher method to have the compiler enforce coverage when enum values change:

```go
type statusHandler struct{}

func (statusHandler) Pending()   {}
func (statusHandler) Approved()  {}
func (statusHandler) Rejected()  {}
func (statusHandler) Completed() {}

var _ validation.StatusMatcher = statusHandler{}

func handleStatus(status validation.Status) error {
	return validation.MatchStatus(status, statusHandler{})
}
```

See the [examples]({{< relref "/examples" >}}) for more complete scenarios.
