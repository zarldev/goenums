package validation

type status int

//go:generate goenums -file status.go
const (
	unknown status = iota
	failed
	passed
	skipped
	scheduled
	running
	booked
)
