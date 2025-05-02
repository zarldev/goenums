package validation

type status int

//go:generate goenums status.go

const (
	failed status = iota // invalid
	passed
	skipped
	scheduled
	running
	booked
)
