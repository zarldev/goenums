package validation

type status int

const (
	unknown status = iota // invalid
	failed
	passed
	skipped
	scheduled
	running
	booked
)
