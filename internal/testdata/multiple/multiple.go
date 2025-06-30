package multipleenums

//go:generate goenums -f multiple.go

type order int

const (
	created     order = iota // CREATED
	approved                 // APPROVED
	processing               // PROCESSING
	readyToShip              // READY_TO_SHIP
	shipped                  // SHIPPED
	delivered                // DELIVERED
	cancelled                // CANCELLED
)

type status int

const (
	failed    status = iota // FAILED
	passed                  // PASSED
	skipped                 // SKIPPED
	scheduled               // SCHEDULED
	running                 // RUNNING
	booked                  // BOOKED
)
