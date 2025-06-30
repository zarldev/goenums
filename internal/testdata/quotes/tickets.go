package tickets

type ticket int // Comment string, Validstate bool

//go:generate goenums -f tickets.go

const (
	unknown           ticket = iota // invalid "Not Found","Missing" "Ticket not found",false
	created                         // "Created Successfully","Created" "Ticket created successfully",true
	pending                         // "In Progress","Pending" "Ticket is being processed",true
	approval_pending                // "Pending Approval","Approval Pending" "Ticket is pending approval",true
	approval_accepted               // "Fully Approved","Approval Accepted" "Ticket has been fully approved",true
	rejected                        // "Has Been Rejected","Rejected" "Ticket has been rejected",false
	completed                       // "Successfully Completed","Completed" "Ticket has been completed",false
)
