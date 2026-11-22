package lenientzero

//go:generate goenums --interfaces=binary state.go

type state int

const (
	below    state = iota - 3
	zero     state = iota - 1
	minusone state = iota - 3 // minusone,""zero""
)
