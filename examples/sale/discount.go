package sale

//go:generate goenums -f discount.go
type discountType int // Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration

const (
	sale       discountType = iota + 1 // false,true,true,false,24*7*time.Hour
	percentage                         // false,false,false,false,24*time.Hour
	amount                             // false,false,false,false,48*time.Hour
	giveaway                           // true,true,false,false,72*time.Hour
)
