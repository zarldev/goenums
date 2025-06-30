package validationstrings

type status int

//go:generate goenums -f status.go

const (
	failed    status = iota // invalid FAILED
	passed                  // PASSED
	skipped                 // SKIPPED
	scheduled               // SCHEDULED
	running                 // RUNNING
	booked                  // BOOKED
)
