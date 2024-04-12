package validation

type status int

const (
	failed status = iota // invalid
	passed
	skipped
	scheduled
	running
	booked
)
