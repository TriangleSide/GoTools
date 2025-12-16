package migration

import "context"

// Manager defines the functions needed to manage and coordinate migrations.
type Manager interface {
	// AcquireDBLock must acquire a database wide lock.
	// It is used in conjunction with EnsureDataStores and ReleaseDBLock.
	AcquireDBLock(ctx context.Context) error

	// EnsureDataStores must ensure the migration data stores (collections, tables, ...) are created.
	// There should be two data stores, one for the migration lock, and one for migration statuses.
	EnsureDataStores(ctx context.Context) error

	// ReleaseDBLock must release the DB lock acquired by AcquireDBLock.
	ReleaseDBLock(ctx context.Context) error

	// AcquireMigrationLock must acquire a migration lock.
	// This is to ensure only one migrator can run at any given time.
	AcquireMigrationLock(ctx context.Context) error

	// MigrationLockHeartbeat is called on a configurable frequency.
	// It is meant to maintain the lock acquired with AcquireMigrationLock.
	MigrationLockHeartbeat(ctx context.Context) error

	// ListStatuses returns data previously stored with PersistStatus.
	ListStatuses(ctx context.Context) ([]PersistedStatus, error)

	// PersistStatus stores or overrides the status of a migration.
	// Order must be unique in the data store.
	PersistStatus(ctx context.Context, order Order, status Status) error

	// ReleaseMigrationLock must release the migration lock acquired with AcquireMigrationLock.
	ReleaseMigrationLock(ctx context.Context) error
}
