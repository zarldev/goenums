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
	return t != unknown
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
