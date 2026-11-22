package contracts

//go:generate goenums --interfaces=binary lenient.go

type mode int

const (
	ready mode = iota + 1
	waiting
)
