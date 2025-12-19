package validation

import (
	"fmt"
	"sync"
)

var (
	// registeredAliases is a map of alias name to its expansion.
	registeredAliases = sync.Map{}
)

// MustRegisterAlias sets the expansion for an alias.
func MustRegisterAlias(name string, expansion string) {
	_, alreadyExists := registeredAliases.LoadOrStore(name, expansion)
	if alreadyExists {
		panic(fmt.Errorf("alias named %s already exists", name))
	}
}

// lookupAlias returns the expansion for an alias if it exists.
func lookupAlias(name string) (string, bool) {
	expansion, exists := registeredAliases.Load(name)
	if !exists {
		return "", false
	}
	return expansion.(string), true
}
