// Code generated by goenums. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums
// using the command:
// goenums planets.go

package solarsystemsimple

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
)

type Planet struct {
	planet
}

type planetsContainer struct {
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

var Planets = planetsContainer{
	MERCURY: Planet{
		planet: mercury,
	},
	VENUS: Planet{
		planet: venus,
	},
	EARTH: Planet{
		planet: earth,
	},
	MARS: Planet{
		planet: mars,
	},
	JUPITER: Planet{
		planet: jupiter,
	},
	SATURN: Planet{
		planet: saturn,
	},
	URANUS: Planet{
		planet: uranus,
	},
	NEPTUNE: Planet{
		planet: neptune,
	},
}

func (c planetsContainer) All() []Planet {
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

func ParsePlanet(a any) (Planet, error) {
	res := invalidPlanet
	switch v := a.(type) {
	case Planet:
		return v, nil
	case []byte:
		res = stringToPlanet(string(v))
	case string:
		res = stringToPlanet(v)
	case fmt.Stringer:
		res = stringToPlanet(v.String())
	case int:
		res = intToPlanet(v)
	case int64:
		res = intToPlanet(int(v))
	case int32:
		res = intToPlanet(int(v))
	}
	return res, nil
}

func stringToPlanet(s string) Planet {
	switch s {
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
	newp, err := ParsePlanet(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p *Planet) Scan(value any) error {
	newp, err := ParsePlanet(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p Planet) Value() (driver.Value, error) {
	return p.String(), nil
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

const _planets_name = "unknownmercuryvenusearthmarsjupitersaturnuranusneptune"

var _planets_index = [...]uint16{0, 7, 14, 19, 24, 28, 35, 41, 47, 54}

func (i planet) String() string {
	if i < 0 || i >= planet(len(_planets_index)-1) {
		return "planets(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _planets_name[_planets_index[i]:_planets_index[i+1]]
}
