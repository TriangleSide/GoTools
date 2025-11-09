package migration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/database/migration"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	t.Run("when an order is registered twice it should fail", func(t *testing.T) {
		t.Parallel()
		reg := migration.NewRegistry()
		registrationOrder := migration.Order(1)
		reggy := &migration.Registration{
			Order:   registrationOrder,
			Migrate: func(ctx context.Context) error { return nil },
			Enabled: true,
		}
		reg.MustRegister(reggy)
		assert.PanicPart(t, func() {
			reg.MustRegister(reggy)
		}, fmt.Sprintf("order %d already exists", registrationOrder))
	})

	t.Run("when the registration fails the validation it should panic", func(t *testing.T) {
		t.Parallel()
		reg := migration.NewRegistry()
		registrationOrder := migration.Order(1)
		assert.PanicPart(t, func() {
			reg.MustRegister(&migration.Registration{
				Order:   registrationOrder,
				Migrate: nil,
				Enabled: true,
			})
		}, "Validation failed for registration")
	})

	t.Run("when OrderedRegistrations is called the registration should be in order", func(t *testing.T) {
		t.Parallel()
		reg := migration.NewRegistry()
		const count = 32
		for i := range count {
			registrationOrder := migration.Order(count - i)
			reg.MustRegister(&migration.Registration{
				Order:   registrationOrder,
				Migrate: func(ctx context.Context) error { return nil },
				Enabled: true,
			})
		}
		ordered := reg.OrderedRegistrations()
		for i := 1; i < len(ordered); i++ {
			assert.True(t, ordered[i-1].Order < ordered[i].Order)
		}
	})
}
