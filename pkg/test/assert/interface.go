package assert

// Testing matches the functions on the testing.T struct.
type Testing interface {
	Helper()
	Error(...any)
	Fatal(...any)
}
