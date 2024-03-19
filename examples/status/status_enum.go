package validation

import (
	"bytes"
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

var Statuses = statusContainer{}

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
