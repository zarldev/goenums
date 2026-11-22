package contracts

//go:generate goenums -f strict.go

type signal int

const (
	below signal = iota - 3
	_
	above   signal = iota + 5
	distant signal = iota + 39
)
