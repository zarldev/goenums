package validation

type status int

//go:generate goenums -f status.go

const (
	failed status = iota // invalid
	passed
	skipped
	scheduled
	running
	booked
)
