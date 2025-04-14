package planets_simple

type planet int

//go:generate goenums -f planets.go
const (
	mercury planet = iota // Mercury
	venus                 // Venus
	earth                 // Earth
	mars                  // Mars
	jupiter               // Jupiter
	saturn                // Saturn
	uranus                // Uranus
	neptune               // Neptune
)
