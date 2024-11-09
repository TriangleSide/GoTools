package logger

import (
	"sync"
)

var (
	lock = sync.RWMutex{}
)
