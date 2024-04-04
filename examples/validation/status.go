package validation

type status int

//go:generate goenums status.go
const (
	unknown status = iota // invalid
	failed
	passed
	skipped
	scheduled
	running
	booked
)
