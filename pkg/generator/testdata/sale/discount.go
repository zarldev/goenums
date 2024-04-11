package sale

//go:generate goenums discount.go
type discountType int

const (
	unknown discountType = iota
	percentage
	amount
	giveaway
)
