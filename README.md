# goenums

goenums is a tool to help you generate go type safe enums that are much more tightly typed than just `iota` defined enums.

# Usage
`goenums -h`
`Usage: goenums <config file path> <output path>`
### Example
Defining the list of enums their type and package in a JSON list is the configuration format.  This allows generating many enums in one shot.
input.json:
```json
{
  "enums": [
    {
      "package": "validation",
      "type": "status",
      "values": [
        "Failed",
        "Passed",
        "Skipped",
        "Scheduled",
        "Running"
      ]
    }
  ]
}
```
Running the following command `goenums ./input.json ./output` will generate `output/validation/status.go` which looks like:

```golang
package validation

import "fmt"

type Status struct {
	status
}

type status int

const (
	unknown status = iota
	failed
	passed
	skipped
	scheduled
	running
)

var (
	strStatusMap = map[status]string{
		failed:    "FAILED",
		passed:    "PASSED",
		skipped:   "SKIPPED",
		scheduled: "SCHEDULED",
		running:   "RUNNING",
	}

	typeStatusMap = map[string]status{
		"FAILED":    failed,
		"PASSED":    passed,
		"SKIPPED":   skipped,
		"SCHEDULED": scheduled,
		"RUNNING":   running,
	}
)

func (t status) String() string {
	return strStatusMap[t]
}

func Parse(a any) Status {
	switch v := a.(type) {
	case Status:
		return v
	case string:
		return Status{stringToStatus(v)}
	case fmt.Stringer:
		return Status{stringToStatus(v.String())}
	case int:
		return Status{status(v)}
	case int64:
		return Status{status(int(v))}
	case int32:
		return Status{status(int(v))}
	}
	return Status{unknown}
}

func stringToStatus(s string) status {
	if v, ok := typeStatusMap[s]; ok {
		return v
	}
	return unknown
}

func (t status) IsValid() bool {
	_, ok := typeStatusMap[t]
    return ok
}

type statussContainer struct {
	UNKNOWN   Status
	FAILED    Status
	PASSED    Status
	SKIPPED   Status
	SCHEDULED Status
	RUNNING   Status
}

var Statuses = statussContainer{
	UNKNOWN:   Status{unknown},
	FAILED:    Status{failed},
	PASSED:    Status{passed},
	SKIPPED:   Status{skipped},
	SCHEDULED: Status{scheduled},
	RUNNING:   Status{running},
}

func (c statussContainer) All() []Status {
	return []Status{
		c.FAILED,
		c.PASSED,
		c.SKIPPED,
		c.SCHEDULED,
		c.RUNNING,
	}
}

func (t Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *Status) UnmarshalJSON(b []byte) error {
	*t = Parse(string(b))
	return nil
}
```
## Features

#### String representation
All enums are generated with a String representation for each enum and JSON Marshaling and UnMarshaling for use in HTTP Request structs.  

#### Extendable
The enums can have additional functionality added by just adding another file in the same package with the extra functionality defined there.  Having the extra functionality in another file will allow the generation and regeneration of the enums to not affect this extra functionality. 

#### Safety
Also the fact that the enums are concrete types with no way to instantiate the nested struct means that you can't just pass the `int` representation of the enum into the `Status` struct.

The above `Validation Status` can be found in the examples directory along with another file extending the behaviour of the `Status` enum and the `config.json` that was used to generate.