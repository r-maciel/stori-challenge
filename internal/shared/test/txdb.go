package test

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DATA-DOG/go-txdb"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	registered  atomic.Bool
	migrateOnce sync.Once
)

// RegisterTxDB registers the txdb driver once, bound to TEST_DATABASE_URL or a sensible default
// pointing to the docker-compose test service.
func RegisterTxDB() (driverName string, baseDSN string) {
	if registered.CompareAndSwap(false, true) {
		baseDSN = os.Getenv("TEST_DATABASE_URL")
		if baseDSN == "" {
			// default to docker-compose db_test service
			baseDSN = "postgres://stori:stori@db_test:5432/stori_test?sslmode=disable"
		}
		txdb.Register("txdb", "pgx", baseDSN)
		return "txdb", baseDSN
	}
	return "txdb", os.Getenv("TEST_DATABASE_URL")
}

// OpenTestDB opens a txdb connection for a single test. Closing the DB rolls back all changes.
func OpenTestDB(testName string) (*sql.DB, error) {
	// Ensure migrations are applied once per test process
	migrateOnce.Do(func() {
		if err := RunDbmateUp(); err != nil {
			panic("db migrations failed: " + err.Error())
		}
	})
	driverName, _ := RegisterTxDB()
	// unique dsn per test connection; txdb uses this to isolate transactions
	dsn := fmt.Sprintf("%s-%d", testName, time.Now().UnixNano())
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	// ensure single connection to keep transaction boundaries tight
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}
