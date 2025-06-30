package solarsystemsimple

type planet int // Gravity float64
//go:generate goenums -f planets.go

const (
	unknown planet = iota // invalid
	mercury               // Mercury 0.378
	venus                 // Venus 0.907
	earth                 // Earth 1
	mars                  // Mars 0.377
	jupiter               // Jupiter 2.36
	saturn                // Saturn 0.916
	uranus                // Uranus 0.889
	neptune               // Neptune 1.12
)
