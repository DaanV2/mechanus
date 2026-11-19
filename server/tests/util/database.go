package util_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DaanV2/mechanus/server/components"
	"github.com/DaanV2/mechanus/server/infrastructure/persistence"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xrand"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gorm.io/gorm/logger"
)

const (
	// EnvTestDatabaseType selects the backend [CreateIntegrationDatabase] uses: memory (default), sqlite or postgres.
	EnvTestDatabaseType = "TEST_DATABASE_TYPE"
	// EnvTestDatabaseDSN is the connection string for the postgres backend, in key=value or URL form.
	EnvTestDatabaseDSN = "TEST_DATABASE_DSN"
)

// CreateDatabase returns a fresh in-memory SQLite database, closed when the spec ends.
// Component tests use this: the database is only setup for them.
func CreateDatabase(setupCtx context.Context) *persistence.DB {
	ginkgo.GinkgoHelper()

	db, err := components.SetupTestDatabase(setupCtx, persistence.WithDBLogger(&GinkgoDBLogger{}))
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), "database setup")
	ginkgo.DeferCleanup(func() {
		gomega.Expect(db.Close(setupCtx)).To(gomega.Succeed(), "database close")
	})

	return db
}

// CreateIntegrationDatabase returns a fresh, empty database on the backend picked by [EnvTestDatabaseType],
// cleaned up when the spec ends. Integration tests use this, so they can run against every supported backend.
func CreateIntegrationDatabase(setupCtx context.Context) *persistence.DB {
	ginkgo.GinkgoHelper()

	switch dbType := persistence.DBType(os.Getenv(EnvTestDatabaseType)); dbType {
	case "", persistence.InMemory:
		return CreateDatabase(setupCtx)
	case persistence.SQLite:
		return createSQLiteDatabase(setupCtx)
	case persistence.PostgreSQL:
		return createPostgresDatabase(setupCtx)
	case persistence.MySQL:
		ginkgo.Fail(EnvTestDatabaseType + ": the integration tests don't support mysql yet")

		return nil
	default:
		ginkgo.Fail(fmt.Sprintf("%s: unsupported database type %q, use memory, sqlite or postgres", EnvTestDatabaseType, dbType))

		return nil
	}
}

// createSQLiteDatabase opens a database file in a temporary directory, removed when the spec ends.
func createSQLiteDatabase(setupCtx context.Context) *persistence.DB {
	ginkgo.GinkgoHelper()

	// The busy timeout makes connections of the pool wait on each other's locks instead of failing
	dsn := filepath.Join(ginkgo.GinkgoT().TempDir(), "test.sqlite") + "?_pragma=busy_timeout(5000)"

	return openDatabase(setupCtx, persistence.WithType(persistence.SQLite), persistence.WithDSN(dsn))
}

// createPostgresDatabase creates a schema of its own for the spec, so specs never see each other's data,
// and drops it when the spec ends.
func createPostgresDatabase(setupCtx context.Context) *persistence.DB {
	ginkgo.GinkgoHelper()

	dsn := os.Getenv(EnvTestDatabaseDSN)
	gomega.Expect(dsn).ToNot(gomega.BeEmpty(), EnvTestDatabaseDSN+" is required for the postgres backend")

	admin, err := persistence.NewDB(persistence.WithType(persistence.PostgreSQL), persistence.WithDSN(dsn), persistence.WithDBLogger(&GinkgoDBLogger{}))
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), "connecting to postgres")

	schema := "test_" + xrand.MustID(16)
	gomega.Expect(admin.WithContext(setupCtx).Exec("CREATE SCHEMA "+schema).Error).To(gomega.Succeed(), "creating schema")
	// Registered first, so it runs after the database below is closed
	ginkgo.DeferCleanup(func(ctx context.Context) {
		gomega.Expect(admin.WithContext(ctx).Exec("DROP SCHEMA "+schema+" CASCADE").Error).To(gomega.Succeed(), "dropping schema")
		gomega.Expect(admin.Close(ctx)).To(gomega.Succeed(), "closing admin connection")
	})

	return openDatabase(setupCtx, persistence.WithType(persistence.PostgreSQL), persistence.WithDSN(withSearchPath(dsn, schema)))
}

// withSearchPath adds the search_path connection parameter to a key=value or URL connection string.
func withSearchPath(dsn, schema string) string {
	if !strings.Contains(dsn, "://") {
		return dsn + " search_path=" + schema
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&search_path=" + schema
	}

	return dsn + "?search_path=" + schema
}

func openDatabase(setupCtx context.Context, opts ...persistence.Option) *persistence.DB {
	ginkgo.GinkgoHelper()

	opts = append(opts, persistence.WithDBLogger(&GinkgoDBLogger{}))
	db, err := components.SetupDatabase(setupCtx, opts...)
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), "database setup")
	ginkgo.DeferCleanup(func(ctx context.Context) {
		gomega.Expect(db.Close(ctx)).To(gomega.Succeed(), "database close")
	})

	return db
}

type GinkgoDBLogger struct{}

// Error implements logger.Interface.
func (g *GinkgoDBLogger) Error(ctx context.Context, msg string, args ...any) {
	ginkgo.GinkgoWriter.Printf("[ERROR]: "+msg, args...)
}

// Warn implements logger.Interface.
func (g *GinkgoDBLogger) Warn(ctx context.Context, msg string, args ...any) {
	ginkgo.GinkgoWriter.Printf("[WARN]: "+msg, args...)
}

// Info implements logger.Interface.
func (g *GinkgoDBLogger) Info(ctx context.Context, msg string, args ...any) {
	ginkgo.GinkgoWriter.Printf("[INFO]: "+msg, args...)
}

// LogMode implements logger.Interface.
func (g *GinkgoDBLogger) LogMode(logger.LogLevel) logger.Interface {
	return g // ignore, its for tests
}

// Trace implements logger.Interface.
func (g *GinkgoDBLogger) Trace(_ context.Context, _ time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()

	if err == nil {
		ginkgo.GinkgoWriter.Printf("[DB QUERY](affected %d): %s\n", rows, sql)
	} else {
		ginkgo.GinkgoWriter.Printf("[DB QUERY](affected %d): %s\n   [ERROR]: %v", rows, sql, err)
	}
}
