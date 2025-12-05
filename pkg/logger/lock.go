package logger

import (
	"sync"
)

var (
	// lock guards access to logger configuration and state.
	lock = sync.RWMutex{}
)
