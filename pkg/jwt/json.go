package jwt

import (
	"encoding/json"
)

var (
	// marshalFunc is used for JSON marshaling and can be overwritten in tests.
	marshalFunc = json.Marshal
)
