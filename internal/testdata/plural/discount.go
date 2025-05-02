package sale

//go:generate goenums -f discount.go
type discountTypes int // Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration

const (
	sale       discountTypes = iota + 1 // false,true,true,false,168h
	percentage                          // false,false,false,false,24h
	amount                              // false,false,false,false,48h
	giveaway                            // true,true,false,false,72h
)
