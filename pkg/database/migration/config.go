package migration

// Config holds parameters for running a migration.
type Config struct {
	// MigrationDeadlineMillis is the maximum time for the migrations to complete.
	MigrationDeadlineMillis int `config:"ENV" config_default:"3600000" validate:"gt=0"`

	// MigrationUnlockDeadlineMillis is the maximum time for a release operation to complete.
	MigrationUnlockDeadlineMillis int `config:"ENV" config_default:"120000" validate:"gt=0"`

	// MigrationHeartbeatIntervalMillis is how often a heart beat is sent to the migration lock.
	MigrationHeartbeatIntervalMillis int `config:"ENV" config_default:"10000" validate:"gt=0"`

	// MigrationHeartbeatFailureRetryCount is how many times to retry the heart beat before quitting.
	MigrationHeartbeatFailureRetryCount int `config:"ENV" config_default:"1" validate:"gte=0"`
}
