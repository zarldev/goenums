package discount

//go:generate goenums -f discount.go
type discountType int // Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration

const (
	sale       discountType = iota + 1 // false,true,true,false,172hr
	percentage                         // false,false,false,false,24hr
	amount                             // false,false,false,false,48hr
	giveaway                           // true,true,false,false,72hr
)
