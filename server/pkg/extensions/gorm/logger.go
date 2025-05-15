package xgorm

import (
	"context"

	"time"

	"github.com/DaanV2/mechanus/server/internal/logging"
	"github.com/charmbracelet/log"
	"gorm.io/gorm/logger"
)

var _ logger.Interface = &GormLogger{}

type GormLogger struct {
	level  log.Level
	logger logging.Enriched
}

func NewGormlogger() *GormLogger {
	result := &GormLogger{
		logger: logging.Enriched{}.WithPrefix("db"),
		level:  log.GetLevel(),
	}

	return result
}

// Error implements logger.Interface.
func (g *GormLogger) Error(ctx context.Context, msg string, args ...any) {
	g.logger.From(ctx).Errorf(msg, args...)
}

// Info implements logger.Interface.
func (g *GormLogger) Info(ctx context.Context, msg string, args ...any) {
	g.logger.From(ctx).Infof(msg, args...)
}

// Warn implements logger.Interface.
func (g *GormLogger) Warn(ctx context.Context, msg string, args ...any) {
	g.logger.From(ctx).Warnf(msg, args...)
}

// LogMode implements logger.Interface.
func (g *GormLogger) LogMode(l logger.LogLevel) logger.Interface {
	switch l {
	case logger.Silent:
		g.level = log.FatalLevel
	case logger.Error:
		g.level = log.ErrorLevel
	case logger.Warn:
		g.level = log.WarnLevel
	case logger.Info:
		g.level = log.InfoLevel
	}

	return g
}

// Trace implements logger.Interface.
func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// If we are warn or higher (fatal) then we only report if we have an error
	if g.level >= log.WarnLevel {
		if err == nil {
			return
		}
	}

	sql, rows := fc()
	logger := g.logger.From(ctx).With("rows", rows, "duration", time.Since(begin))
	if err != nil {
		logger = logger.With("err", err)
	}
	logger.Debug("[Query]: " + sql)
}
