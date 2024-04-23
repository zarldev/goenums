// Code generated by goenums. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums
// using the command:
// goenums testdata/validation-strings/status.go

package validationstrings

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
	FAILED    Status
	PASSED    Status
	SKIPPED   Status
	SCHEDULED Status
	RUNNING   Status
	BOOKED    Status
}

var Statuses = statusesContainer{
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
	case "FAILED":
		return Statuses.FAILED
	case "PASSED":
		return Statuses.PASSED
	case "SKIPPED":
		return Statuses.SKIPPED
	case "SCHEDULED":
		return Statuses.SCHEDULED
	case "RUNNING":
		return Statuses.RUNNING
	case "BOOKED":
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
	_ = x[failed-0]
	_ = x[passed-1]
	_ = x[skipped-2]
	_ = x[scheduled-3]
	_ = x[running-4]
	_ = x[booked-5]
}

const _statuses_name = "FAILEDPASSEDSKIPPEDSCHEDULEDRUNNINGBOOKED"

var _statuses_index = [...]uint16{0, 6, 12, 19, 28, 35, 41}

func (i status) String() string {
	if i < 0 || i >= status(len(_statuses_index)-1) {
		return "statuses(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _statuses_name[_statuses_index[i]:_statuses_index[i+1]]
}
