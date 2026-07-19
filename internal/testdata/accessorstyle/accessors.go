package accessorstyle

type status int

//go:generate goenums --accessor-style=go accessors.go

const (
	pendingApproval status = iota
	apiError
	http2Enabled
	userID
)
