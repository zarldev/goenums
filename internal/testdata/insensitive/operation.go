package insensitive

type operationType int // Active bool

//go:generate goenums -i -f operation.go

const (
	update operationType = iota + 1 // true
	remove                          // false
)
