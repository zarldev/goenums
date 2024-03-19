package milkywaysimple

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Planet struct {
	planet
}

type planetContainer struct {
	UNKNOWN Planet
	MERCURY Planet
	VENUS   Planet
	EARTH   Planet
	MARS    Planet
	JUPITER Planet
	SATURN  Planet
	URANUS  Planet
	NEPTUNE Planet
}

var Planets = planetContainer{}

func (c planetContainer) All() []Planet {
	return []Planet{
		c.MERCURY,
		c.VENUS,
		c.EARTH,
		c.MARS,
		c.JUPITER,
		c.SATURN,
		c.URANUS,
		c.NEPTUNE,
	}
}

var invalidPlanet = Planet{}

func ParsePlanet(a any) Planet {
	switch v := a.(type) {
	case Planet:
		return v
	case string:
		return stringToPlanet(v)
	case fmt.Stringer:
		return stringToPlanet(v.String())
	case int:
		return intToPlanet(v)
	case int64:
		return intToPlanet(int(v))
	case int32:
		return intToPlanet(int(v))
	}
	return invalidPlanet
}

func stringToPlanet(s string) Planet {
	lwr := strings.ToLower(s)
	switch lwr {
	case "unknown":
		return Planets.UNKNOWN
	case "mercury":
		return Planets.MERCURY
	case "venus":
		return Planets.VENUS
	case "earth":
		return Planets.EARTH
	case "mars":
		return Planets.MARS
	case "jupiter":
		return Planets.JUPITER
	case "saturn":
		return Planets.SATURN
	case "uranus":
		return Planets.URANUS
	case "neptune":
		return Planets.NEPTUNE
	}
	return invalidPlanet
}

func intToPlanet(i int) Planet {
	if i < 0 || i >= len(Planets.All()) {
		return invalidPlanet
	}
	return Planets.All()[i]
}

func ExhaustivePlanets(f func(Planet)) {
	for _, p := range Planets.All() {
		f(p)
	}
}

var validPlanets = map[Planet]bool{
	Planets.MERCURY: true,
	Planets.VENUS:   true,
	Planets.EARTH:   true,
	Planets.MARS:    true,
	Planets.JUPITER: true,
	Planets.SATURN:  true,
	Planets.URANUS:  true,
	Planets.NEPTUNE: true,
}

func (p Planet) IsValid() bool {
	return validPlanets[p]
}

func (p Planet) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Planet) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, `"`), ` `)
	*p = ParsePlanet(string(b))
	return nil
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the goenums command to generate them again.
	// Does not identify newly added constant values unless order changes
	var x [1]struct{}
	_ = x[unknown-0]
	_ = x[mercury-1]
	_ = x[venus-2]
	_ = x[earth-3]
	_ = x[mars-4]
	_ = x[jupiter-5]
	_ = x[saturn-6]
	_ = x[uranus-7]
	_ = x[neptune-8]
}

const _planet_name = "unknownmercuryvenusearthmarsjupitersaturnuranusneptune"

var _planet_index = [...]uint16{0, 7, 14, 19, 24, 28, 35, 41, 47, 54}

func (i planet) String() string {
	if i < 0 || i >= planet(len(_planet_index)-1) {
		return "planet(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _planet_name[_planet_index[i]:_planet_index[i+1]]
}
