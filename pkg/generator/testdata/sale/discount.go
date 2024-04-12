package sale

//go:generate goenums discount.go
type discountType int

const (
	sale discountType = iota + 1
	percentage
	amount
	giveaway
)
