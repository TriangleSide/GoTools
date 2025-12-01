package migration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

// Status represents the status of a persisted migration.
type Status string

const (
	Pending   Status = "PENDING"
	Started   Status = "STARTED"
	Failed    Status = "FAILED"
	Completed Status = "COMPLETED"
)

// PersistedStatus is the data stored in the migration table.
type PersistedStatus struct {
	Order  Order  `validate:"gte=0"`
	Status Status `validate:"oneof=PENDING STARTED FAILED COMPLETED"`
}

// Manager defines the functions needed to manage and coordinate migrations.
type Manager interface {
	// AcquireDBLock must acquire a database wide lock.
	// It is used in conjunction with EnsureDataStores and ReleaseDBLock.
	AcquireDBLock(context.Context) error

	// EnsureDataStores must ensure the migration data stores (collections, tables, ...) are created.
	// There should be two data stores, one for the migration lock, and one for migration statuses.
	EnsureDataStores(context.Context) error

	// ReleaseDBLock must release the DB lock acquired by AcquireDBLock.
	ReleaseDBLock(context.Context) error

	// AcquireMigrationLock must acquire a migration lock.
	// This is to ensure only one migrator can run at any given time.
	AcquireMigrationLock(context.Context) error

	// MigrationLockHeartbeat is called on a configurable frequency.
	// It is meant to maintain the lock acquired with AcquireMigrationLock.
	MigrationLockHeartbeat(context.Context) error

	// ListStatuses returns data previously stored with PersistStatus.
	ListStatuses(context.Context) ([]PersistedStatus, error)

	// PersistStatus stores or overrides the status of a migration.
	// Order must be unique in the data store.
	PersistStatus(context.Context, Order, Status) error

	// ReleaseMigrationLock must release the migration lock acquired with AcquireMigrationLock.
	ReleaseMigrationLock(context.Context) error
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

	var migrationsToRun []*Registration
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

// listMigrationsToRun compares the registered migrations to the persisted statuses.
// It returns the list of migrations that need to be run.
func listMigrationsToRun(ctx context.Context, manager Manager, reg *Registry) ([]*Registration, error) {
	orderToPersistedStatus, err := fetchPersistedStatuses(ctx, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the persisted statuses (%w)", err)
	}

	latestCompletedMigration := Order(-1)
	migrationsToRun := make([]*Registration, 0)

	for _, registeredMigration := range reg.OrderedRegistrations() {
		migrationStatus, migrationStatusFound := orderToPersistedStatus[registeredMigration.Order]

		if !registeredMigration.Enabled {
			if migrationStatusFound {
				logger.Warnf("Migration with order %d is disabled but previously run with status %s. Skipping.",
					registeredMigration.Order, migrationStatus)
				delete(orderToPersistedStatus, registeredMigration.Order)
			} else {
				logger.Debugf("Migration with order %d is disabled and not previously run. Skipping.", registeredMigration.Order)
			}
			continue
		}

		if migrationStatusFound {
			delete(orderToPersistedStatus, registeredMigration.Order)
			if migrationStatus == Completed {
				logger.Debugf("Registration with order %d already completed. Skipping.", registeredMigration.Order)
				if registeredMigration.Order > latestCompletedMigration {
					latestCompletedMigration = registeredMigration.Order
				}
			} else {
				logger.Debugf("Will attempt to run the migration with order %d and status %s again.", registeredMigration.Order, migrationStatus)
				migrationsToRun = append(migrationsToRun, registeredMigration)
			}
		} else {
			logger.Debugf("New migration with order %d found.", registeredMigration.Order)
			migrationsToRun = append(migrationsToRun, registeredMigration)
		}
	}

	if len(orderToPersistedStatus) != 0 {
		return nil, fmt.Errorf("found persisted migration(s) that are not in the registry (%+v)", orderToPersistedStatus)
	}

	for _, migrationToRun := range migrationsToRun {
		if migrationToRun.Order < latestCompletedMigration {
			return nil, fmt.Errorf("cannot run migrations out of order (found %d but latest completed is %d)", migrationToRun.Order, latestCompletedMigration)
		}
	}

	return migrationsToRun, nil
}

// runMigrations first persists the statuses of all the migrations as PENDING.
// Then it attempts to run the migrations while keeping the statuses updated.
func runMigrations(ctx context.Context, migrationsToRun []*Registration, manager Manager) error {
	for _, registered := range migrationsToRun {
		if err := manager.PersistStatus(ctx, registered.Order, Pending); err != nil {
			return fmt.Errorf("failed to persist the status %s for the migration order %d (%w)", Pending, registered.Order, err)
		}
	}

	for _, migrationToRun := range migrationsToRun {
		migrationCtx, logEntry := logger.AddField(ctx, "order", migrationToRun.Order)
		logEntry.Debug("Starting migration.")
		startTime := time.Now()
		if err := manager.PersistStatus(migrationCtx, migrationToRun.Order, Started); err != nil {
			return fmt.Errorf("failed to persist the status %s for the migration order %d (%w)", Started, migrationToRun.Order, err)
		}
		if err := migrationToRun.Migrate(migrationCtx); err != nil {
			err = fmt.Errorf("failed to complete the migration with order %d (%w)", migrationToRun.Order, err)
			if failedStatusErr := manager.PersistStatus(migrationCtx, migrationToRun.Order, Failed); failedStatusErr != nil {
				return fmt.Errorf("%w and failed to persist its status to %s (%w)", err, Failed, failedStatusErr)
			}
			return err
		}
		if err := manager.PersistStatus(migrationCtx, migrationToRun.Order, Completed); err != nil {
			return fmt.Errorf("failed to persist the status %s for the migration order %d (%w)", Completed, migrationToRun.Order, err)
		}
		logEntry.Debugf("Migration finished in %s.", time.Since(startTime))
	}

	return nil
}
