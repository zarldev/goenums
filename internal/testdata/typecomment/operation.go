package typecomment

// A descriptive type comment that is not a field declaration must not be
// parsed as a field (regression: it previously emitted an uncompilable
// "<nil>" struct field). The enum should generate as a simple field-less enum.
type operationType int //operations

//go:generate goenums -f operation.go

const (
	update operationType = iota + 1
	remove
)
