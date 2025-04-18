package invalid

type invalid int // invalid invalid floaf

const (
	invalid invalid = iota // invalid
	valid           = iota // invalid
	another         = iota // invalid



var invalidInvalid = invalid.invalid
var invalidValid = invalid.valid