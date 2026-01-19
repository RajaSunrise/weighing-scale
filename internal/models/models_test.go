package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/stretchr/testify/assert"
)

func TestWeighingRecordIndex(t *testing.T) {
	// Setup in-memory SQLite DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate
	err = db.AutoMigrate(&WeighingRecord{})
	assert.NoError(t, err)

	// Verify Index Exists
	migrator := db.Migrator()

	// GORM creates index name based on table and column usually,
	// but HasIndex can check by field name if we pass the struct.
	// Actually HasIndex second arg is the index name.
	// But `migrator.HasIndex(&WeighingRecord{}, "idx_weighing_records_weighed_at")` is the explicit check.
	// Let's rely on GORM checking if it has an index on that column.

	// Let's check specifically for the index on the column 'weighed_at'
	hasIndex := migrator.HasIndex(&WeighingRecord{}, "WeighedAt")
	assert.True(t, hasIndex, "Index on WeighedAt should exist")
}
