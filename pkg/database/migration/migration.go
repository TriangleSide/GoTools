package migration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TriangleSide/GoTools/pkg/validation"
)

// Status represents the status of a persisted migration.
type Status string

const (
	// Pending indicates a migration is queued to run.
	Pending Status = "PENDING"

	// Started indicates a migration is currently running.
	Started Status = "STARTED"

	// Failed indicates a migration failed during execution.
	Failed Status = "FAILED"

	// Completed indicates a migration successfully finished.
	Completed Status = "COMPLETED"
)

// PersistedStatus is the data stored in the migration table.
type PersistedStatus struct {
	Order  Order  `validate:"gte=0"`
	Status Status `validate:"oneof=PENDING STARTED FAILED COMPLETED"`
}

// Migrate orchestrates database migrations using the provided Manager and options.
func Migrate(manager Manager, opts ...Option) (returnErr error) {
	migrateCfg := configure(opts...)
	cfg, err := migrateCfg.configProvider()
	if err != nil {
		return fmt.Errorf("failed to get the migration configuration (%w)", err)
	}
	reg := migrateCfg.registry
	if reg == nil {
		return errors.New("migration registry is nil")
	}

	var releaseMigrationLockErr error = nil
	releaseMigrationLockWG := sync.WaitGroup{}

	ctxDeadline := time.Now().Add(time.Millisecond * time.Duration(cfg.MigrationDeadlineMillis))
	ctx, cancel := context.WithDeadline(context.Background(), ctxDeadline)
	defer func() {
		cancel()
		releaseMigrationLockWG.Wait()
		if releaseMigrationLockErr != nil {
			returnErr = errors.Join(returnErr, releaseMigrationLockErr)
		}
	}()

	if err = ensureDataStores(ctx, manager, cfg); err != nil {
		return fmt.Errorf("failed to ensure the data stores are created (%w)", err)
	}

	if err = manager.AcquireMigrationLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire the migration lock (%w)", err)
	}

	releaseMigrationLockWG.Go(func() {
		if releaseMigrationLockErr = heartbeatAndRelease(ctx, manager, cfg); releaseMigrationLockErr != nil {
			cancel()
		}
	})

	var migrationsToRun []migrationToRun
	if migrationsToRun, err = listMigrationsToRun(ctx, manager, reg); err != nil {
		return fmt.Errorf("failed to list the migrations to run (%w)", err)
	}

	if err = runMigrations(ctx, migrationsToRun, manager); err != nil {
		return fmt.Errorf("error while running migrations (%w)", err)
	}

	return nil
}

// ensureDataStores acquires a DB lock, creates the migration data stores, then releases the DB lock.
func ensureDataStores(ctx context.Context, manager Manager, cfg *Config) (returnErr error) {
	if err := manager.AcquireDBLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire the database lock (%w)", err)
	}

	defer func() {
		releaseDeadline := time.Now().Add(time.Millisecond * time.Duration(cfg.MigrationUnlockDeadlineMillis))
		releaseCtx, releaseCancel := context.WithDeadline(context.WithoutCancel(ctx), releaseDeadline)
		defer releaseCancel()
		if releaseErr := manager.ReleaseDBLock(releaseCtx); releaseErr != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("failed to release the database lock (%w)", releaseErr))
		}
	}()

	if err := manager.EnsureDataStores(ctx); err != nil {
		return fmt.Errorf("failed to ensure the data stores are created (%w)", err)
	}

	return nil
}

// heartbeatAndRelease calls MigrationLockHeartbeat on a configured frequency.
// Once the context is canceled, it calls ReleaseMigrationLock.
func heartbeatAndRelease(ctx context.Context, manager Manager, cfg *Config) (returnErr error) {
	defer func() {
		releaseDeadline := time.Now().Add(time.Millisecond * time.Duration(cfg.MigrationUnlockDeadlineMillis))
		releaseCtx, releaseCancel := context.WithDeadline(context.WithoutCancel(ctx), releaseDeadline)
		defer releaseCancel()
		if releaseErr := manager.ReleaseMigrationLock(releaseCtx); releaseErr != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("failed to release the migration lock (%w)", releaseErr))
		}
	}()

	heartbeatInterval := time.Millisecond * time.Duration(cfg.MigrationHeartbeatIntervalMillis)
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()

	var heartbeatErr error = nil
	var successiveHeartbeatErrCount = 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeatTicker.C:
			if heartbeatErr = manager.MigrationLockHeartbeat(ctx); heartbeatErr != nil {
				successiveHeartbeatErrCount++
			} else {
				successiveHeartbeatErrCount = 0
			}
		}
		if successiveHeartbeatErrCount > cfg.MigrationHeartbeatFailureRetryCount {
			return fmt.Errorf("heartbeat failed %d time(s) with latest error of (%w)", successiveHeartbeatErrCount, heartbeatErr)
		}
	}
}

// fetchPersistedStatuses retrieves the persisted statuses.
func fetchPersistedStatuses(ctx context.Context, manager Manager) (map[Order]Status, error) {
	persistedStatuses, err := manager.ListStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list the persisted statuses (%w)", err)
	}
	orderToPersistedStatus := make(map[Order]Status)
	for _, persistedStatus := range persistedStatuses {
		if err := validation.Struct(persistedStatus); err != nil {
			return nil, fmt.Errorf("failed while validating the persisted status (%w)", err)
		}
		if _, alreadyFound := orderToPersistedStatus[persistedStatus.Order]; alreadyFound {
			return nil, fmt.Errorf("found two persisted statuses with order %d", persistedStatus.Order)
		}
		orderToPersistedStatus[persistedStatus.Order] = persistedStatus.Status
	}
	return orderToPersistedStatus, nil
}

// migrationToRun holds a registration and its previous persisted status.
type migrationToRun struct {
	registration   *Registration
	previousStatus Status
}

// listMigrationsToRun compares the registered migrations to the persisted statuses.
// It returns the list of migrations that need to be run along with their previous status.
func listMigrationsToRun(ctx context.Context, manager Manager, reg *Registry) ([]migrationToRun, error) {
	orderToPersistedStatus, err := fetchPersistedStatuses(ctx, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the persisted statuses (%w)", err)
	}

	orderedMigrations := reg.OrderedRegistrations()
	if err := validatePersistedMigrationsAreRegistered(orderToPersistedStatus, orderedMigrations); err != nil {
		return nil, err
	}
	latestCompletedMigration := latestCompletedMigrationOrder(orderToPersistedStatus, orderedMigrations)
	migrationsToRun := enabledMigrationsThatArentCompleted(orderToPersistedStatus, orderedMigrations)
	if err := validateMigrationsInOrder(migrationsToRun, latestCompletedMigration); err != nil {
		return nil, err
	}

	return migrationsToRun, nil
}

// validatePersistedMigrationsAreRegistered ensures no persisted migration order is missing from the registry.
func validatePersistedMigrationsAreRegistered(orderToPersistedStatus map[Order]Status, orderedMigrations []*Registration) error {
	registeredMigrationOrders := make(map[Order]struct{}, len(orderedMigrations))
	for _, registeredMigration := range orderedMigrations {
		registeredMigrationOrders[registeredMigration.Order] = struct{}{}
	}

	for persistedOrder := range orderToPersistedStatus {
		if _, found := registeredMigrationOrders[persistedOrder]; !found {
			return fmt.Errorf("found persisted migration(s) that are not in the registry (%+v)", orderToPersistedStatus)
		}
	}

	return nil
}

// latestCompletedMigrationOrder returns the greatest registered order with a persisted COMPLETED status, or -1 if none exist.
func latestCompletedMigrationOrder(orderToPersistedStatus map[Order]Status, orderedMigrations []*Registration) Order {
	latestCompletedMigration := Order(-1)
	for _, registeredMigration := range orderedMigrations {
		migrationStatus, migrationStatusFound := orderToPersistedStatus[registeredMigration.Order]
		if migrationStatusFound && migrationStatus == Completed && registeredMigration.Order > latestCompletedMigration {
			latestCompletedMigration = registeredMigration.Order
		}
	}
	return latestCompletedMigration
}

// enabledMigrationsThatArentCompleted returns enabled migrations that are not persisted as COMPLETED, with their previous status.
func enabledMigrationsThatArentCompleted(orderToPersistedStatus map[Order]Status, orderedMigrations []*Registration) []migrationToRun {
	migrationsToRun := make([]migrationToRun, 0)
	for _, registeredMigration := range orderedMigrations {
		if !registeredMigration.Enabled {
			continue
		}
		migrationStatus, found := orderToPersistedStatus[registeredMigration.Order]
		if migrationStatus == Completed {
			continue
		}
		previousStatus := Pending
		if found {
			previousStatus = migrationStatus
		}
		migrationsToRun = append(migrationsToRun, migrationToRun{
			registration:   registeredMigration,
			previousStatus: previousStatus,
		})
	}
	return migrationsToRun
}

// validateMigrationsInOrder ensures no migration will be run with an order less than the latest completed order.
func validateMigrationsInOrder(migrationsToRun []migrationToRun, latestCompletedMigration Order) error {
	for _, mtr := range migrationsToRun {
		if mtr.registration.Order < latestCompletedMigration {
			return fmt.Errorf("cannot run migrations out of order (found %d but latest completed is %d)", mtr.registration.Order, latestCompletedMigration)
		}
	}
	return nil
}

// runMigrations first persists the statuses of all the migrations as PENDING.
// Then it attempts to run the migrations while keeping the statuses updated.
func runMigrations(ctx context.Context, migrationsToRun []migrationToRun, manager Manager) error {
	for _, mtr := range migrationsToRun {
		if err := manager.PersistStatus(ctx, mtr.registration.Order, Pending); err != nil {
			return fmt.Errorf("failed to persist the status %s for the migration order %d (%w)", Pending, mtr.registration.Order, err)
		}
	}

	for _, mtr := range migrationsToRun {
		if err := manager.PersistStatus(ctx, mtr.registration.Order, Started); err != nil {
			return fmt.Errorf("failed to persist the status %s for the migration order %d (%w)", Started, mtr.registration.Order, err)
		}
		if err := mtr.registration.Migrate(ctx, mtr.previousStatus); err != nil {
			err = fmt.Errorf("failed to complete the migration with order %d (%w)", mtr.registration.Order, err)
			if failedStatusErr := manager.PersistStatus(ctx, mtr.registration.Order, Failed); failedStatusErr != nil {
				return fmt.Errorf("%w and failed to persist its status to %s (%w)", err, Failed, failedStatusErr)
			}
			return err
		}
		if err := manager.PersistStatus(ctx, mtr.registration.Order, Completed); err != nil {
			return fmt.Errorf("failed to persist the status %s for the migration order %d (%w)", Completed, mtr.registration.Order, err)
		}
	}

	return nil
}
