package errors

// Error is the standard JSON response an API endpoint makes when an error occurs in the endpoint handler.
type Error struct {
	Message string `json:"message"`
}
