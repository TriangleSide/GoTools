package migration

import (
	"context"
	"fmt"
	"sort"

	"github.com/TriangleSide/GoTools/pkg/validation"
)

// Order represents the sequence in which migrations are meant to be run.
type Order int

// Registration defines a migration callback and its order.
type Registration struct {
	// Order is compared to other migrations to determine its run sequence.
	Order Order `validate:"gte=0"`

	// Migrate is invoked to run the migration.
	// This function MUST be idempotent and retryable.
	// The Status parameter contains the previous persisted status, defaulting to Pending if none exists.
	Migrate func(context.Context, Status) error `validate:"required"`

	// Enabled indicates if this migration is to be run or not.
	// A migration could be disabled if another migration covers it.
	Enabled bool
}

// Registry stores migration registrations keyed by their order.
type Registry struct {
	registrations map[Order]*Registration
}

// NewRegistry returns a new empty migration registry.
func NewRegistry() *Registry {
	return &Registry{
		registrations: make(map[Order]*Registration),
	}
}

// MustRegister stores a migration registration in the registry.
func (r *Registry) MustRegister(registration *Registration) {
	if err := validation.Struct(registration); err != nil {
		panic(fmt.Sprintf("Validation failed for registration (%s).", err.Error()))
	}
	if _, ok := r.registrations[registration.Order]; ok {
		panic(fmt.Sprintf("Registration with order %d already exists.", registration.Order))
	}
	r.registrations[registration.Order] = registration
}

// OrderedRegistrations returns an ordered list of the registrations in the registry.
// The registrations are sorted by their Order.
func (r *Registry) OrderedRegistrations() []*Registration {
	ordered := make([]*Registration, 0, len(r.registrations))
	for _, registration := range r.registrations {
		ordered = append(ordered, registration)
	}

	sort.Slice(ordered, func(a, b int) bool {
		return ordered[a].Order < ordered[b].Order
	})

	return ordered
}
