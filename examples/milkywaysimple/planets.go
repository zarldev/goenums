package milkywaysimple

type planet int

//go:generate goenums -file planets.go
const (
	unknown planet = iota // invalid
	mercury
	venus
	earth
	mars
	jupiter
	saturn
	uranus
	neptune
)
