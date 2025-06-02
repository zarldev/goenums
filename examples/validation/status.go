package validation

type status int

//go:generate goenums -f -c status.go

const (
	unknown   status = iota // invalid UNKNOWN
	failed                  // FAILED
	passed                  // PASSED
	skipped                 // SKIPPED
	scheduled               // SCHEDULED
	running                 // RUNNING
	booked                  // BOOKED
)
