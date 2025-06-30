package crypto

//go:generate goenums -c -f negative.go

type algorithm int

const (
	None algorithm = iota - 1 // invalid
	AES256
	ChaCha20
)
