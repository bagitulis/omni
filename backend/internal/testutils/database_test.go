package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestSetupTeardown verifies the postgres container lifecycle works correctly
func TestSetupTeardown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	// This test verifies that SetupTestPostgres starts the container successfully
	db := SetupTestPostgres(t)
	require.NotNil(t, db, "database connection should not be nil")

	// Verify we can connect and execute a simple query
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	require.NoError(t, err, "failed to execute simple query")
	assert.Equal(t, 1, result, "simple query should return 1")
}

// TestGormConnection verifies that GORM works with the testcontainer
func TestGormConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	db := SetupTestPostgres(t)
	require.NotNil(t, db, "database connection should not be nil")

	// Create a simple test table
	type TestModel struct {
		ID   int `gorm:"primaryKey"`
		Name string
	}

	// Auto-migrate
	err := db.AutoMigrate(&TestModel{})
	require.NoError(t, err, "failed to auto-migrate test model")

	// Test CREATE
	testRecord := TestModel{ID: 1, Name: "test_record"}
	err = db.Create(&testRecord).Error
	require.NoError(t, err, "failed to create test record")

	// Test READ
	var readRecord TestModel
	err = db.First(&readRecord, 1).Error
	require.NoError(t, err, "failed to read test record")
	assert.Equal(t, "test_record", readRecord.Name, "record name should match")

	// Test UPDATE
	err = db.Model(&readRecord).Update("name", "updated_record").Error
	require.NoError(t, err, "failed to update test record")

	// Verify update
	var updatedRecord TestModel
	err = db.First(&updatedRecord, 1).Error
	require.NoError(t, err, "failed to read updated record")
	assert.Equal(t, "updated_record", updatedRecord.Name, "updated name should match")

	// Test DELETE
	err = db.Delete(&TestModel{}, 1).Error
	require.NoError(t, err, "failed to delete test record")

	// Verify deletion
	var deletedRecord TestModel
	result := db.First(&deletedRecord, 1)
	assert.True(t, result.Error == gorm.ErrRecordNotFound, "record should be deleted")
}

// TestMultipleDatabases verifies multiple concurrent containers work independently
func TestMultipleDatabases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	// Start two separate database containers
	db1 := SetupTestPostgres(t)
	require.NotNil(t, db1, "first database connection should not be nil")

	db2 := SetupTestPostgres(t)
	require.NotNil(t, db2, "second database connection should not be nil")

	// Verify both can execute queries independently
	var result1, result2 int
	err1 := db1.Raw("SELECT 42").Scan(&result1).Error
	err2 := db2.Raw("SELECT 99").Scan(&result2).Error

	require.NoError(t, err1, "first database query should succeed")
	require.NoError(t, err2, "second database query should succeed")

	assert.Equal(t, 42, result1, "first database should return correct value")
	assert.Equal(t, 99, result2, "second database should return correct value")
}

// TestContextPropagation verifies GORM context operations work
func TestContextPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	db := SetupTestPostgres(t)
	require.NotNil(t, db, "database connection should not be nil")

	ctx := context.Background()

	// Test that WithContext works
	var result int
	err := db.WithContext(ctx).Raw("SELECT 100").Scan(&result).Error
	require.NoError(t, err, "context propagation should work")
	assert.Equal(t, 100, result, "query with context should return correct value")
}
