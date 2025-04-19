package planets_gravity_only

type planet int // Gravity[float64]

//go:generate goenums -f planets.go
const (
	mercury planet = iota // 2
	venus                 // 2
	earth                 // 1
	mars                  // 3
	jupiter               // 69
	saturn                // 58
	uranus                // 25
	neptune               // 24
)
