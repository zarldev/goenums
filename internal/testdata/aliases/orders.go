package orders

type order int

//go:generate goenums -f orders.go

const (
	created     order = iota // CREATED
	approved                 // APPROVED
	processing               // PROCESSING
	readyToShip              // READY_TO_SHIP,ready-to-ship,ready_to_ship
	shipped                  // SHIPPED
	delivered                // DELIVERED
	cancelled                // CANCELLED
)
