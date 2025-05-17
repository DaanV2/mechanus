package util_test

import (
	"context"
	"time"

	"github.com/DaanV2/mechanus/server/internal/components"
	"github.com/DaanV2/mechanus/server/pkg/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"
)

func CreateDatabase() *database.DB {
	GinkgoHelper()

	db, err := components.SetupTestDatabase(database.WithDBLogger(&GinkgoDBLogger{}))
	Expect(err).ToNot(HaveOccurred())

	return db
}

type GinkgoDBLogger struct{}

// Error implements logger.Interface.
func (g *GinkgoDBLogger) Error(ctx context.Context, msg string, args ...any) {
	GinkgoWriter.Printf("[ERROR]: " + msg, args...)
}

// Warn implements logger.Interface.
func (g *GinkgoDBLogger) Warn(ctx context.Context, msg string, args ...any) {
	GinkgoWriter.Printf("[WARN]: " + msg, args...)
}

// Info implements logger.Interface.
func (g *GinkgoDBLogger) Info(ctx context.Context, msg string, args ...any) {
	GinkgoWriter.Printf("[INFO]: " + msg, args...)
}

// LogMode implements logger.Interface.
func (g *GinkgoDBLogger) LogMode(logger.LogLevel) logger.Interface {
	return g //ignore, its for tests
}

// Trace implements logger.Interface.
func (g *GinkgoDBLogger) Trace(_ context.Context, _ time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()

	if err == nil {
		GinkgoWriter.Printf("[DB QUERY](affected %d): %s\n", rows, sql)
	} else {
		GinkgoWriter.Printf("[DB QUERY](affected %d): %s\n   [ERROR]: %v", rows, sql, err)
	}
}
