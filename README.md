# goenums

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![build](https://github.com/zarldev/goenums/actions/workflows/go.yml/badge.svg)

goenums is a tool to help you generate go type safe enums that are much more tightly typed than just `iota` defined enums.

# Installation
`go install github.com/zarldev/goenums@latest`

# Usage
```
Usage of goenums:
  -file string
        Path to the file to generate enums from
  -valuer string
        The return value type of db valuer implementation, support int and string (default "string")
```

### Example
Defining the list of enums in the respective go file and then point the goenum binary at the require file.  This can be specified in the go generate command like below:
For example we have the file below called status.go :

```golang
package validation

type status int

//go:generate goenums -file status.go
const (
	unknown status = iota
	failed
	passed
	skipped
	scheduled
	running
	booked
)
```
Now running the `go generate` command will generate the following code in a new file called `status_enum.go`

```golang
package validation

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type Status struct {
	status
}

type statusContainer struct {
	UNKNOWN   Status
	FAILED    Status
	PASSED    Status
	SKIPPED   Status
	SCHEDULED Status
	RUNNING   Status
	BOOKED    Status
}

var Statuses = statusContainer{
	UNKNOWN: Status{
		status: unknown,
	},
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

func (c statusContainer) All() []Status {
	return []Status{
		c.UNKNOWN,
		c.FAILED,
		c.PASSED,
		c.SKIPPED,
		c.SCHEDULED,
		c.RUNNING,
		c.BOOKED,
	}
}

var invalidStatus = Status{}

func ParseStatus(a any) Status {
	switch v := a.(type) {
	case Status:
		return v
	case []byte:
		return stringToStatus(string(v))
	case string:
		return stringToStatus(v)
	case fmt.Stringer:
		return stringToStatus(v.String())
	case int:
		return intToStatus(v)
	case int64:
		return intToStatus(int(v))
	case int32:
		return intToStatus(int(v))
	}
	return invalidStatus
}

func stringToStatus(s string) Status {
	lwr := strings.ToLower(s)
	switch lwr {
	case "unknown":
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
	Statuses.UNKNOWN:   true,
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
	*p = ParseStatus(string(b))
	return nil
}

func (p *Status) Scan(value any) error {
	*p = ParseStatus(value)
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

const _status_name = "unknownfailedpassedskippedscheduledrunningbooked"

var _status_index = [...]uint16{0, 7, 13, 19, 26, 35, 42, 48}

func (i status) String() string {
	if i < 0 || i >= status(len(_status_index)-1) {
		return "status(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _status_name[_status_index[i]:_status_index[i+1]]
}

```
## Features

#### String representation
All enums are generated with a String representation for each enum and JSON Marshaling and UnMarshaling for use in HTTP Request structs.  The string function is now the same as the `go cmd stringer` for the base case.

#### Extendable
The enums can have additional functionality added by just adding comments to the type definition and corresponding values to the comments in the iota definitions.  There is also the `invalid` comment flag which will no longer include the value in the exhaustive list. 

For example we have the file below called planets.go :

```golang
package milkywaysimple

import (
	"bytes"
	"database/sql/driver"
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

var Planets = planetContainer{
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
	case []byte:
		return stringToPlanet(string(v))
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

func (p *Planet) Scan(value any) error {
	*p = ParsePlanet(value)
	return nil
}

func (p Planet) Value() (driver.Value, error) {
	return int64(p.planet), nil
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
```

With the above code generated we can use the `ExhaustivePlanets` to iterate over all Enums for example:

```golang
package main

import (
	"fmt"

	"github.com/zarldev/goenums/examples/milkyway"
)

func main() {
	weightKg := 100.0
	milkyway.ExhaustivePlanets(func(p milkyway.Planet) {
		// calculate weight on each planet
		gravity := p.Gravity
		planetMass := weightKg * gravity
		fmt.Printf("Weight on %s is %fKg with gravity %f\n", p, planetMass, gravity)
	})
}
```

#### Safety
Also the fact that the enums are concrete types with no way to instantiate the nested struct means that you can't just pass the `int` representation of the enum into the generated wrapper struct.

The above `Status` and `Planet` examples can be found in the examples directory.

### Mentions
[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)
