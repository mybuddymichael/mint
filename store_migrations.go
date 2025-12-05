package main

// CurrentSchemaVersion is the latest schema version
const CurrentSchemaVersion = 1

// Migration is a function that modifies the store schema
type Migration func(*Store) error

// migrations is a sequential list of migration functions
// Index corresponds to "from version" (migrations[0] upgrades v0->v1, etc.)
var migrations = []Migration{
	migration_1, // v0 -> v1: Add schema_version field
}

// RunMigrations runs all pending migrations on a store
// Returns true if any migrations were run
func RunMigrations(store *Store) (bool, error) {
	startVersion := store.SchemaVersion

	// Run all migrations from current version to latest
	for store.SchemaVersion < CurrentSchemaVersion {
		migrationIndex := store.SchemaVersion
		if migrationIndex >= len(migrations) {
			// Should never happen, but defensive
			break
		}

		if err := migrations[migrationIndex](store); err != nil {
			return false, err
		}

		store.SchemaVersion++
	}

	return store.SchemaVersion > startVersion, nil
}

// migration_1 adds schema_version field
// This is a no-op migration since the field is already added to the struct
// and will be set by the RunMigrations -> SchemaVersion++ flow
func migration_1(store *Store) error {
	// No-op: schema_version field already added to struct
	// Version will be incremented after this function
	return nil
}
