package status

// Code represents the status of a span.
type Code int

const (
	// Unset is the default status code indicating no status has been set.
	Unset Code = iota

	// Error indicates the operation tracked by the span failed.
	Error

	// Success indicates the operation tracked by the span succeeded.
	Success
)

// String returns a string representation of the status code.
func (c Code) String() string {
	switch c {
	case Error:
		return "Error"
	case Success:
		return "Success"
	default:
		return "Unset"
	}
}
