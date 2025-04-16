package validationstrings

type status int

const (
	failed    status = iota // invalid FAILED
	passed                  // PASSED
	skipped                 // SKIPPED
	scheduled               // SCHEDULED
	running                 // RUNNING
	booked                  // BOOKED
)
