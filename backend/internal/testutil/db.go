package testutil

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbCounter atomic.Int64

// NewTestDB opens a SQLite in-memory database, auto-migrates the given models,
// and returns a *gorm.DB ready for testing. Each call creates a fresh isolated
// database to prevent cross-test data leakage. The database is automatically
// closed when the test finishes.
func NewTestDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	// Each test gets its own unique in-memory database to avoid shared cache contamination.
	id := dbCounter.Add(1)
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:testdb_%s_%d?mode=memory&cache=shared", safeName, id)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("testutil: open sqlite: %v", err)
	}

	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("testutil: auto migrate: %v", err)
		}
	}

	// Close the underlying sql.DB when the test finishes.
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}
