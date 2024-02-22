package cmd

import "fmt"

type Operation struct {
	operation
}

type operation int

const (
	unknown operation = iota
	escalated
	archived
	deleted
	completed
)

var (
	strOperationMap = map[operation]string{
		escalated: "ESCALATED",
		archived:  "ARCHIVED",
		deleted:   "DELETED",
		completed: "COMPLETED",
	}

	typeOperationMap = map[string]operation{
		"ESCALATED": escalated,
		"ARCHIVED":  archived,
		"DELETED":   deleted,
		"COMPLETED": completed,
	}
)

func (t operation) String() string {
	return strOperationMap[t]
}

func Parse(a any) Operation {
	switch v := a.(type) {
	case Operation:
		return v
	case string:
		return Operation{stringToOperation(v)}
	case fmt.Stringer:
		return Operation{stringToOperation(v.String())}
	case int:
		return Operation{operation(v)}
	case int64:
		return Operation{operation(int(v))}
	case int32:
		return Operation{operation(int(v))}
	}
	return Operation{unknown}
}

func stringToOperation(s string) operation {
	if v, ok := typeOperationMap[s]; ok {
		return v
	}
	return unknown
}

func (t operation) IsValid() bool {
	return t != unknown
}

type operationsContainer struct {
	UNKNOWN   Operation
	ESCALATED Operation
	ARCHIVED  Operation
	DELETED   Operation
	COMPLETED Operation
}

var Operations = operationsContainer{
	UNKNOWN:   Operation{unknown},
	ESCALATED: Operation{escalated},
	ARCHIVED:  Operation{archived},
	DELETED:   Operation{deleted},
	COMPLETED: Operation{completed},
}

func (c operationsContainer) All() []Operation {
	return []Operation{
		c.ESCALATED,
		c.ARCHIVED,
		c.DELETED,
		c.COMPLETED,
	}
}

func (t Operation) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *Operation) UnmarshalJSON(b []byte) error {
	*t = Parse(string(b))
	return nil
}
