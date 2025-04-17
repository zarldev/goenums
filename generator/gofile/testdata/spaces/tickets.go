package spaces

type ticketStatus int

//go:generate goenums status.go
const (
	unknown   ticketStatus = iota // invalid "Not Found"
	pending                       // "In Progress"
	approved                      // "Fully Approved"
	rejected                      // "Has Been Rejected"
	completed                     // "Successfully Completed"
)
