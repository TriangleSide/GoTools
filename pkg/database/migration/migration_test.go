package migration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/database/migration"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

type managerRecorder struct {
	Operations          []string
	PersistedMigrations []migration.PersistedStatus

	Heartbeat            chan struct{}
	HeartbeatErrors      []error
	HeartbeatCount       int
	MigrationUnlockCount int
	FailOnStatus         string

	AcquireDBLockError          error
	EnsureDataStoresError       error
	ReleaseDBLockError          error
	MigrationLockError          error
	MigrationLockHeartbeatError error
	ListStatusesError           error
	PersistStatusError          error
	ReleaseMigrationLockError   error
}

func (r *managerRecorder) AcquireDBLock(ctx context.Context) error {
	r.Operations = append(r.Operations, "AcquireDBLock()")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return r.AcquireDBLockError
}

func (r *managerRecorder) EnsureDataStores(ctx context.Context) error {
	r.Operations = append(r.Operations, "EnsureDataStores()")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return r.EnsureDataStoresError
}

func (r *managerRecorder) ReleaseDBLock(ctx context.Context) error {
	r.Operations = append(r.Operations, "ReleaseDBLock()")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return r.ReleaseDBLockError
}

func (r *managerRecorder) AcquireMigrationLock(ctx context.Context) error {
	r.Operations = append(r.Operations, "AcquireMigrationLock()")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return r.MigrationLockError
}

func (r *managerRecorder) MigrationLockHeartbeat(ctx context.Context) error {
	r.HeartbeatCount++
	if r.Heartbeat != nil {
		r.Heartbeat <- struct{}{}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(r.HeartbeatErrors) > 0 {
		index := r.HeartbeatCount - 1
		if index < len(r.HeartbeatErrors) {
			return r.HeartbeatErrors[index]
		}
	}
	return r.MigrationLockHeartbeatError
}

func (r *managerRecorder) ListStatuses(ctx context.Context) ([]migration.PersistedStatus, error) {
	r.Operations = append(r.Operations, "ListStatuses()")
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return r.PersistedMigrations, r.ListStatusesError
}

func (r *managerRecorder) PersistStatus(ctx context.Context, order migration.Order, status migration.Status) error {
	r.Operations = append(r.Operations, fmt.Sprintf("PersistStatus(order=%d, status=%s)", order, status))
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if string(status) == r.FailOnStatus {
		return errors.New("fail on " + string(status))
	}
	return r.PersistStatusError
}

func (r *managerRecorder) ReleaseMigrationLock(ctx context.Context) error {
	r.MigrationUnlockCount++
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return r.ReleaseMigrationLockError
}

func standardRegisteredMigration(manager *managerRecorder, order migration.Order) *migration.Registration {
	return &migration.Registration{
		Order: order,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			manager.Operations = append(manager.Operations, fmt.Sprintf("Migration%d.Migrate()", order))
			return ctx.Err()
		},
		Enabled: true,
	}
}

func TestMigrate_ConfigProviderFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	reg := migration.NewRegistry()
	opts := []migration.Option{
		migration.WithConfigProvider(func() (*migration.Config, error) {
			return nil, errors.New("configProvider error")
		}),
		migration.WithRegistry(reg),
	}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to get the migration configuration")
	var expectedOps []string
	assert.Equals(t, expectedOps, manager.Operations)
}

func TestMigrate_Success_RunsMigrationsSuccessfully(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(2)))
	reg.MustRegister(&migration.Registration{
		Order: 3,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			manager.Operations = append(manager.Operations, "Migration3.Migrate()")
			return ctx.Err()
		},
		Enabled: false,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=2, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
		"PersistStatus(order=2, status=STARTED)",
		"Migration2.Migrate()",
		"PersistStatus(order=2, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_AcquireDBLockFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		AcquireDBLockError: errors.New("AcquireDBLock error"),
	}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to acquire the database lock (AcquireDBLock error)")
	expectedOps := []string{
		"AcquireDBLock()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 0)
}

func TestMigrate_EnsureDataStoresFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		EnsureDataStoresError: errors.New("EnsureDataStores error"),
	}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to ensure the data stores are created (EnsureDataStores error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 0)
}

func TestMigrate_EnsureDataStoresAndReleaseDBLockFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		EnsureDataStoresError: errors.New("EnsureDataStores error"),
		ReleaseDBLockError:    errors.New("ReleaseDBLockError error"),
	}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to ensure the data stores are created (EnsureDataStores error)")
	assert.ErrorPart(t, err, "failed to release the database lock (ReleaseDBLockError error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 0)
}

func TestMigrate_ReleaseDBLockFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		ReleaseDBLockError: errors.New("ReleaseDBLock error"),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to release the database lock (ReleaseDBLock error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 0)
}

func TestMigrate_AcquireMigrationLockFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		MigrationLockError: errors.New("AcquireMigrationLock error"),
	}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to acquire the migration lock (AcquireMigrationLock error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 0)
}

func TestMigrate_ListStatusesFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		ListStatusesError: errors.New("ListStatuses error"),
	}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to list the persisted statuses (ListStatuses error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_InvalidPersistedStatus_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: "INVALID"},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "the value is not one of the allowed values")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_FailedMigrations_RetriesThem(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Failed},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_StartedMigrations_RetriesThem(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Started},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_PendingMigrations_RetriesThem(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Pending},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_MigrateFunctionFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(context.Context, migration.Status) error {
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return errors.New("migrate error")
		},
		Enabled: true,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to complete the migration with order 1 (migrate error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=FAILED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_MigrateFunctionFailsAndPersistFailedStatusFails_ReturnsBothErrors(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		FailOnStatus: string(migration.Failed),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(_ context.Context, _ migration.Status) error {
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return errors.New("migrate error")
		},
		Enabled: true,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to complete the migration with order 1 (migrate error)")
	assert.ErrorPart(t, err, "failed to persist its status to FAILED (fail on FAILED)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=FAILED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_PersistStatusPendingFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		FailOnStatus: string(migration.Pending),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to persist the status PENDING for the migration order 1 (fail on PENDING)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_PersistedMigrationsNotInRegistry_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Completed},
		},
	}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "found persisted migration(s) that are not in the registry")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_DisabledRegistrationWithPersistedStatus_SkipsMigration(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Completed},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return ctx.Err()
		},
		Enabled: false,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_ReleaseMigrationLockFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		ReleaseMigrationLockError: errors.New("ReleaseMigrationLock error"),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to release the migration lock (ReleaseMigrationLock error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_HeartbeatFailsContinuously_CancelsContextAndStopsMigrations(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		MigrationLockHeartbeatError: errors.New("heartbeat error"),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			<-ctx.Done()
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return ctx.Err()
		},
		Enabled: true,
	})
	opts := []migration.Option{
		migration.WithConfigProvider(func() (*migration.Config, error) {
			cfg, _ := config.Process[migration.Config]()
			cfg.MigrationHeartbeatIntervalMillis = 1
			cfg.MigrationHeartbeatFailureRetryCount = 2
			return cfg, nil
		}),
		migration.WithRegistry(reg),
	}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to complete the migration with order 1 (context canceled)")
	assert.ErrorPart(t, err, "failed to persist its status to FAILED (context canceled))")
	assert.ErrorPart(t, err, "heartbeat failed 3 time(s) with latest error of (heartbeat error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=FAILED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.HeartbeatCount, 3)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_HeartbeatAndReleaseMigrationLockFail_CancelsContextAndStopsMigrations(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		MigrationLockHeartbeatError: errors.New("heartbeat error"),
		ReleaseMigrationLockError:   errors.New("ReleaseMigrationLockError error"),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			<-ctx.Done()
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return ctx.Err()
		},
		Enabled: true,
	})
	opts := []migration.Option{
		migration.WithConfigProvider(func() (*migration.Config, error) {
			cfg, _ := config.Process[migration.Config]()
			cfg.MigrationHeartbeatIntervalMillis = 1
			cfg.MigrationHeartbeatFailureRetryCount = 2
			return cfg, nil
		}),
		migration.WithRegistry(reg),
	}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to complete the migration with order 1 (context canceled)")
	assert.ErrorPart(t, err, "failed to persist its status to FAILED (context canceled))")
	assert.ErrorPart(t, err, "heartbeat failed 3 time(s) with latest error of (heartbeat error)")
	assert.ErrorPart(t, err, "failed to release the migration lock (ReleaseMigrationLockError error)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=FAILED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.HeartbeatCount, 3)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_HeartbeatSucceeds_DoesNotPreventProgress(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		Heartbeat: make(chan struct{}),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			<-manager.Heartbeat
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return ctx.Err()
		},
		Enabled: true,
	})
	opts := []migration.Option{
		migration.WithConfigProvider(func() (*migration.Config, error) {
			cfg, _ := config.Process[migration.Config]()
			cfg.MigrationHeartbeatIntervalMillis = 10
			return cfg, nil
		}),
		migration.WithRegistry(reg),
	}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.True(t, manager.HeartbeatCount > 0)
}

func TestMigrate_HeartbeatRecoversAfterError_ResetsSuccessiveFailures(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		Heartbeat:       make(chan struct{}),
		HeartbeatErrors: []error{errors.New("heartbeat error"), nil, errors.New("heartbeat error")},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			heartbeatDone := make(chan struct{})
			go func() {
				count := 0
				doneClosed := false
				closeDone := func() {
					if !doneClosed {
						close(heartbeatDone)
						doneClosed = true
					}
				}
				for {
					select {
					case <-ctx.Done():
						closeDone()
						return
					case <-manager.Heartbeat:
						if count < 3 {
							count++
							if count == 3 {
								closeDone()
							}
						}
					}
				}
			}()
			select {
			case <-heartbeatDone:
			case <-ctx.Done():
				return ctx.Err()
			}
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return ctx.Err()
		},
		Enabled: true,
	})
	opts := []migration.Option{
		migration.WithConfigProvider(func() (*migration.Config, error) {
			cfg, _ := config.Process[migration.Config]()
			cfg.MigrationHeartbeatIntervalMillis = 1
			cfg.MigrationHeartbeatFailureRetryCount = 1
			return cfg, nil
		}),
		migration.WithRegistry(reg),
	}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.True(t, manager.HeartbeatCount >= 3)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_DuplicatePersistedStatusOrder_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Completed},
			{Order: 1, Status: migration.Completed},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "found two persisted statuses with order 1")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_CompletedStatus_SkipsMigration(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Completed},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_PersistStatusStartedFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		FailOnStatus: string(migration.Started),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to persist the status STARTED for the migration order 1 (fail on STARTED)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_PersistStatusCompletedFails_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		FailOnStatus: string(migration.Completed),
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "failed to persist the status COMPLETED for the migration order 1 (fail on COMPLETED)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate()",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_OutOfSequencePersistedStatus_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 2, Status: migration.Completed},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(2)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.ErrorPart(t, err, "cannot run migrations out of order (found 1 but latest completed is 2)")
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_NilRegistry_ReturnsError(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	err := migration.Migrate(manager, migration.WithRegistry(nil))
	assert.ErrorPart(t, err, "registry is nil")
	err = migration.Migrate(manager)
	assert.ErrorPart(t, err, "registry is nil")
}

func TestMigrate_EmptyRegistry_Succeeds(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	reg := migration.NewRegistry()
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_DisabledMigrationWithoutPersistedStatus_SkipsMigration(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(ctx context.Context, _ migration.Status) error {
			manager.Operations = append(manager.Operations, "Migration1.Migrate()")
			return ctx.Err()
		},
		Enabled: false,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestMigrate_MultipleCompletedMigrationsWithNewMigration_RunsNewMigration(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Completed},
			{Order: 2, Status: migration.Completed},
		},
	}
	reg := migration.NewRegistry()
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(1)))
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(2)))
	reg.MustRegister(standardRegisteredMigration(manager, migration.Order(3)))
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=3, status=PENDING)",
		"PersistStatus(order=3, status=STARTED)",
		"Migration3.Migrate()",
		"PersistStatus(order=3, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
	assert.Equals(t, manager.MigrationUnlockCount, 1)
}

func TestPersistedStatus_ValidStatuses_PassesValidation(t *testing.T) {
	t.Parallel()
	assert.NoError(t, validation.Struct(migration.PersistedStatus{Order: 0, Status: migration.Pending}))
	assert.NoError(t, validation.Struct(migration.PersistedStatus{Order: 1, Status: migration.Started}))
	assert.NoError(t, validation.Struct(migration.PersistedStatus{Order: 2, Status: migration.Failed}))
	assert.NoError(t, validation.Struct(migration.PersistedStatus{Order: 3, Status: migration.Completed}))
}

func TestPersistedStatus_InvalidStatuses_FailsValidation(t *testing.T) {
	t.Parallel()
	assert.Error(t, validation.Struct(migration.PersistedStatus{Order: 4, Status: ""}))
	assert.Error(t, validation.Struct(migration.PersistedStatus{Order: 5, Status: "UNKNOWN"}))
	assert.Error(t, validation.Struct(migration.PersistedStatus{Order: -1, Status: migration.Pending}))
}

func TestMigrate_NoPersistedStatus_PassesPendingAsDefault(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{}
	var receivedStatus migration.Status
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(_ context.Context, previousStatus migration.Status) error {
			receivedStatus = previousStatus
			manager.Operations = append(manager.Operations, fmt.Sprintf("Migration1.Migrate(previousStatus=%s)", previousStatus))
			return nil
		},
		Enabled: true,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	assert.Equals(t, migration.Pending, receivedStatus)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate(previousStatus=PENDING)",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
}

func TestMigrate_PersistedStatusFailed_PassesFailedToPreviousStatus(t *testing.T) {
	t.Parallel()
	manager := &managerRecorder{
		PersistedMigrations: []migration.PersistedStatus{
			{Order: 1, Status: migration.Failed},
		},
	}
	var receivedStatus migration.Status
	reg := migration.NewRegistry()
	reg.MustRegister(&migration.Registration{
		Order: 1,
		Migrate: func(_ context.Context, previousStatus migration.Status) error {
			receivedStatus = previousStatus
			manager.Operations = append(manager.Operations, fmt.Sprintf("Migration1.Migrate(previousStatus=%s)", previousStatus))
			return nil
		},
		Enabled: true,
	})
	opts := []migration.Option{migration.WithRegistry(reg)}
	err := migration.Migrate(manager, opts...)
	assert.NoError(t, err)
	assert.Equals(t, migration.Failed, receivedStatus)
	expectedOps := []string{
		"AcquireDBLock()",
		"EnsureDataStores()",
		"ReleaseDBLock()",
		"AcquireMigrationLock()",
		"ListStatuses()",
		"PersistStatus(order=1, status=PENDING)",
		"PersistStatus(order=1, status=STARTED)",
		"Migration1.Migrate(previousStatus=FAILED)",
		"PersistStatus(order=1, status=COMPLETED)",
	}
	assert.Equals(t, expectedOps, manager.Operations)
}
