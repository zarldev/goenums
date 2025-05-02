package noenums

//go:generate goenums noenums.go

// Not a proper enum (no type)
const (
	First  = 1
	Second = 2
)

type mixed int

const (
	One   mixed = 1
	Two   mixed = 2
	Three       = "three" // Not using the enum type
)
