package migration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/database/migration"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
)

func TestMustRegister_DuplicateOrder_Panics(t *testing.T) {
	t.Parallel()
	reg := migration.NewRegistry()
	registrationOrder := migration.Order(1)
	reggy := &migration.Registration{
		Order:   registrationOrder,
		Migrate: func(context.Context, migration.Status) error { return nil },
		Enabled: true,
	}
	reg.MustRegister(reggy)
	assert.PanicPart(t, func() {
		reg.MustRegister(reggy)
	}, fmt.Sprintf("order %d already exists", registrationOrder))
}

func TestMustRegister_ValidationFails_Panics(t *testing.T) {
	t.Parallel()
	reg := migration.NewRegistry()
	registrationOrder := migration.Order(1)
	assert.PanicPart(t, func() {
		reg.MustRegister(&migration.Registration{
			Order:   registrationOrder,
			Migrate: nil,
			Enabled: true,
		})
	}, "validation failed for registration")
}

func TestOrderedRegistrations_MultipleRegistrations_ReturnsInOrder(t *testing.T) {
	t.Parallel()
	reg := migration.NewRegistry()
	const count = 32
	for i := range count {
		registrationOrder := migration.Order(count - i)
		reg.MustRegister(&migration.Registration{
			Order:   registrationOrder,
			Migrate: func(context.Context, migration.Status) error { return nil },
			Enabled: true,
		})
	}
	ordered := reg.OrderedRegistrations()
	for i := 1; i < len(ordered); i++ {
		assert.True(t, ordered[i-1].Order < ordered[i].Order)
	}
}
