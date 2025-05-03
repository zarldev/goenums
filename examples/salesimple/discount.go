package sale

//go:generate goenums -f discount.go
type discountType int // Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration

const (
	sale       discountType = iota + 1 // false,true,true,false,172h
	percentage                         // false,false,false,false,24h
	amount                             // false,false,false,false,48h
	giveaway                           // true,true,false,false,72h
)
