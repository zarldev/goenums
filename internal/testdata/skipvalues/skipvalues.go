package skipvalues

type version int

//go:generate goenums -f skipvalues.go
const (
	V1 version = iota + 1
	_
	V3
	V4
	_
	_
	V7
)
