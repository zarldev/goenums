# goenums

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![build](https://github.com/zarldev/goenums/actions/workflows/go.yml/badge.svg)

goenums is a tool to help you generate go type safe enums that are much more tightly typed than just `iota` defined enums.

# Installation
```
go install github.com/zarldev/goenums@latest
```

# Usage
```
$ goenums -h
   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/
Usage: goenums [options] filename
Options:
  -f
  -failfast
        Enable failfast mode - fail on generation of invalid enum while parsing (default: false)
  -h
  -help
        Print help information
  -v
  -version
        Print version information
```

### Example
Defining the list of enums in the respective go file and then point the goenum binary at the require file.  This can be specified in the go generate command like below:
For example we have the file below called status.go :

```golang
package validation

type status int

//go:generate goenums status.go
const (
	unknown status = iota // invalid
	failed
	passed
	skipped
	scheduled
	running
	booked
)
```
Now running the `go generate` command will generate the following code in a new file called `statuses_enums.go`

```golang
// Code generated by goenums. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums
// using the command:
// goenums status.go

package validation

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
)

type Status struct {
	status
}

type statusesContainer struct {
	UNKNOWN   Status
	FAILED    Status
	PASSED    Status
	SKIPPED   Status
	SCHEDULED Status
	RUNNING   Status
	BOOKED    Status
}

var Statuses = statusesContainer{
	FAILED: Status{
		status: failed,
	},
	PASSED: Status{
		status: passed,
	},
	SKIPPED: Status{
		status: skipped,
	},
	SCHEDULED: Status{
		status: scheduled,
	},
	RUNNING: Status{
		status: running,
	},
	BOOKED: Status{
		status: booked,
	},
}

func (c statusesContainer) All() []Status {
	return []Status{
		c.FAILED,
		c.PASSED,
		c.SKIPPED,
		c.SCHEDULED,
		c.RUNNING,
		c.BOOKED,
	}
}

var invalidStatus = Status{}

func ParseStatus(a any) (Status, error) {
	res := invalidStatus
	switch v := a.(type) {
	case Status:
		return v, nil
	case []byte:
		res = stringToStatus(string(v))
	case string:
		res = stringToStatus(v)
	case fmt.Stringer:
		res = stringToStatus(v.String())
	case int:
		res = intToStatus(v)
	case int64:
		res = intToStatus(int(v))
	case int32:
		res = intToStatus(int(v))
	}
	return res, nil
}

func stringToStatus(s string) Status {
	switch s {
	case "invalid":
		return Statuses.UNKNOWN
	case "failed":
		return Statuses.FAILED
	case "passed":
		return Statuses.PASSED
	case "skipped":
		return Statuses.SKIPPED
	case "scheduled":
		return Statuses.SCHEDULED
	case "running":
		return Statuses.RUNNING
	case "booked":
		return Statuses.BOOKED
	}
	return invalidStatus
}

func intToStatus(i int) Status {
	if i < 0 || i >= len(Statuses.All()) {
		return invalidStatus
	}
	return Statuses.All()[i]
}

func ExhaustiveStatuss(f func(Status)) {
	for _, p := range Statuses.All() {
		f(p)
	}
}

var validStatuses = map[Status]bool{
	Statuses.FAILED:    true,
	Statuses.PASSED:    true,
	Statuses.SKIPPED:   true,
	Statuses.SCHEDULED: true,
	Statuses.RUNNING:   true,
	Statuses.BOOKED:    true,
}

func (p Status) IsValid() bool {
	return validStatuses[p]
}

func (p Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Status) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, `"`), ` `)
	newp, err := ParseStatus(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p *Status) Scan(value any) error {
	newp, err := ParseStatus(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p Status) Value() (driver.Value, error) {
	return p.String(), nil
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the goenums command to generate them again.
	// Does not identify newly added constant values unless order changes
	var x [1]struct{}
	_ = x[unknown-0]
	_ = x[failed-1]
	_ = x[passed-2]
	_ = x[skipped-3]
	_ = x[scheduled-4]
	_ = x[running-5]
	_ = x[booked-6]
}

const _statuses_name = "invalidfailedpassedskippedscheduledrunningbooked"

var _statuses_index = [...]uint16{0, 7, 13, 19, 26, 35, 42, 48}

func (i status) String() string {
	if i < 0 || i >= status(len(_statuses_index)-1) {
		return "statuses(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _statuses_name[_statuses_index[i]:_statuses_index[i+1]]
}
```
## Features

#### String representation
All enums are generated with a String representation for each enum and JSON Marshaling and UnMarshaling for use in HTTP Request structs.  The string function is now the same as the `go cmd stringer` for the base case.

#### JSON & Database Storage
The generated enum type also implements the JSON.UnMarshal, JSON.Marshal interfaces along with the sql.Scanner and sql.Valuer interface to handle parsing over the wire via HTTP or a Database.

##### Error On Invalid
You can enable the generator to adjust the `JSONUnmarshal` method so that it will return an error if an enum is found to be invalid.
This is triggered by the failfast flag `-f` or `-failfast`. 

Here is the updated ParseXXX function for the `Status` example where we have enabled failfast in the go generate command.

```golang
//go:generate goenums -f status.go
type status int

func ParseStatus(a any) (Status, error) {
	res := invalidStatus
	switch v := a.(type) {
	case Status:
		return v, nil
	case []byte:
		res = stringToStatus(string(v))
	case string:
		res = stringToStatus(v)
	case fmt.Stringer:
		res = stringToStatus(v.String())
	case int:
		res = intToStatus(v)
	case int64:
		res = intToStatus(int(v))
	case int32:
		res = intToStatus(int(v))
	}
	if res == invalidStatus {
		return res, fmt.Errorf("failed to parse %v", a)
	}
	return res, nil
}
```

#### Extendable
The enums can have additional functionality added by just adding comments to the type definition and corresponding values to the comments in the iota definitions.  There is also the `invalid` comment flag which will no longer include the value in the exhaustive list.

Extensions via comments is a comma separated list of `Name` and `Type` declarations, these declarations can be done in 1 of 3 formats depending on preference.

1. Spaces `Gravity float64,RadiusKm float64,MassKg float64,OrbitKm float64`
2. Square Brackets `Gravity[float64],RadiusKm[float64],MassKg[float64],OrbitKm[float64]`
3. Parenthesis `Gravity(float64),RadiusKm(float64),MassKg(float64),OrbitKm(float64)`

For example we have the file below called planets.go :

```golang
package solarsystem

type planet int // Gravity[float64],RadiusKm[float64],MassKg[float64],OrbitKm[float64],OrbitDays[float64],SurfacePressureBars[float64],Moons[int],Rings[bool]

//go:generate goenums planets.go
const (
	unknown planet = iota // invalid
	mercury               // Mercury 0.378,2439.7,3.3e23,57910000,88,0.0000000001,0,false
	venus                 // Venus 0.907,6051.8,4.87e24,108200000,225,92,0,false
	earth                 // Earth 1,6378.1,5.97e24,149600000,365,1,1,false
	mars                  // Mars 0.377,3389.5,6.42e23,227900000,687,0.01,2,false
	jupiter               // Jupiter 2.36,69911,1.90e27,778600000,4333,20,4,true
	saturn                // Saturn 0.916,58232,5.68e26,1433500000,10759,1,7,true
	uranus                // Uranus 0.889,25362,8.68e25,2872500000,30687,1.3,13,true
	neptune               // Neptune 1.12,24622,1.02e26,4495100000,60190,1.5,2,true
)
```

Now running the `go generate` command will generate the following code in a new file called `planets_enums.go`
```golang
// Code generated by goenums. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums
// using the command:
// goenums planets.go

package solarsystem

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
)

type Planet struct {
	planet
	Gravity             float64
	RadiusKm            float64
	MassKg              float64
	OrbitKm             float64
	OrbitDays           float64
	SurfacePressureBars float64
	Moons               int
	Rings               bool
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
		planet:              mercury,
		Gravity:             0.378,
		RadiusKm:            2439.7,
		MassKg:              3.3e23,
		OrbitKm:             57910000,
		OrbitDays:           88,
		SurfacePressureBars: 0.0000000001,
		Moons:               0,
		Rings:               false,
	},
	VENUS: Planet{
		planet:              venus,
		Gravity:             0.907,
		RadiusKm:            6051.8,
		MassKg:              4.87e24,
		OrbitKm:             108200000,
		OrbitDays:           225,
		SurfacePressureBars: 92,
		Moons:               0,
		Rings:               false,
	},
	EARTH: Planet{
		planet:              earth,
		Gravity:             1,
		RadiusKm:            6378.1,
		MassKg:              5.97e24,
		OrbitKm:             149600000,
		OrbitDays:           365,
		SurfacePressureBars: 1,
		Moons:               1,
		Rings:               false,
	},
	MARS: Planet{
		planet:              mars,
		Gravity:             0.377,
		RadiusKm:            3389.5,
		MassKg:              6.42e23,
		OrbitKm:             227900000,
		OrbitDays:           687,
		SurfacePressureBars: 0.01,
		Moons:               2,
		Rings:               false,
	},
	JUPITER: Planet{
		planet:              jupiter,
		Gravity:             2.36,
		RadiusKm:            69911,
		MassKg:              1.90e27,
		OrbitKm:             778600000,
		OrbitDays:           4333,
		SurfacePressureBars: 20,
		Moons:               4,
		Rings:               true,
	},
	SATURN: Planet{
		planet:              saturn,
		Gravity:             0.916,
		RadiusKm:            58232,
		MassKg:              5.68e26,
		OrbitKm:             1433500000,
		OrbitDays:           10759,
		SurfacePressureBars: 1,
		Moons:               7,
		Rings:               true,
	},
	URANUS: Planet{
		planet:              uranus,
		Gravity:             0.889,
		RadiusKm:            25362,
		MassKg:              8.68e25,
		OrbitKm:             2872500000,
		OrbitDays:           30687,
		SurfacePressureBars: 1.3,
		Moons:               13,
		Rings:               true,
	},
	NEPTUNE: Planet{
		planet:              neptune,
		Gravity:             1.12,
		RadiusKm:            24622,
		MassKg:              1.02e26,
		OrbitKm:             4495100000,
		OrbitDays:           60190,
		SurfacePressureBars: 1.5,
		Moons:               2,
		Rings:               true,
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
	case "invalid":
		return Planets.UNKNOWN
	case "Mercury":
		return Planets.MERCURY
	case "Venus":
		return Planets.VENUS
	case "Earth":
		return Planets.EARTH
	case "Mars":
		return Planets.MARS
	case "Jupiter":
		return Planets.JUPITER
	case "Saturn":
		return Planets.SATURN
	case "Uranus":
		return Planets.URANUS
	case "Neptune":
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

const _planets_name = "invalidMercuryVenusEarthMarsJupiterSaturnUranusNeptune"

var _planets_index = [...]uint16{0, 7, 14, 19, 24, 28, 35, 41, 47, 54}

func (i planet) String() string {
	if i < 0 || i >= planet(len(_planets_index)-1) {
		return "planets(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _planets_name[_planets_index[i]:_planets_index[i+1]]
}
```

With the above code generated we can use the `ExhaustivePlanets` to iterate over all Enums for example:

```golang
package main

import (
	"fmt"

	"github.com/zarldev/goenums/examples/solarsystem"
)

func main() {
	weightKg := 100.0
	solarsystem.ExhaustivePlanets(func(p solarsystem.Planet) {
		// calculate weight on each planet
		gravity := p.Gravity
		planetMass := weightKg * gravity
		fmt.Printf("Weight on %s is %fKg with gravity %f\n", p, planetMass, gravity)
	})
}
```

#### Safety
Also the fact that the enums are concrete types with no way to instantiate the nested struct means that you can't just pass the `int` representation of the enum into the generated wrapper struct.

The above `Status` and `Planet` examples can be found in the examples directory.  There is also a `DiscountType` example to show handling of camelCase formatted input enums.

### Mentions
[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)
