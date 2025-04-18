package sale

//go:generate goenums -f sale.go
type sale int // Available bool, Started bool, Finished bool, Cancelled bool, Duration time.Duration

const (
	sales      sale = iota + 1 // false,true,true,false,24*7*time.Hour
	percentage                 // false,false,false,false,24*time.Hour
	amount                     // false,false,false,false,48*time.Hour
	giveaway                   // true,true,false,false,72*time.Hour
)
