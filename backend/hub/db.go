package hub

import (
	"context"
	"log"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

var models = []any{&Tool{}, &CommandLineTool{}, &ServiceTool{}, &HTTPTool{}, &CallingLog{}, &Dependency{}, &ConcurrencyGroup{}}

// InitDB initializes the database connection and performs auto migration for all models.
func InitDB(ctx context.Context, isProduction bool) {
	var (
		err           error
		SlowThreshold = 50 * time.Millisecond
		LogLevel      = logger.Info
	)
	if isProduction {
		SlowThreshold = 200 * time.Millisecond
		LogLevel = logger.Error
	}

	db, err = gorm.Open(sqlite.Open("hub.db"), &gorm.Config{
		Logger: logger.New(log.New(log.Writer(), "\r\n", log.LstdFlags|log.Lmicroseconds|log.Lshortfile), logger.Config{
			SlowThreshold:             SlowThreshold,
			LogLevel:                  LogLevel,
			IgnoreRecordNotFoundError: isProduction,
			ParameterizedQueries:      !isProduction,
			Colorful:                  !isProduction,
		}),
	})
	if err != nil {
		runtime.LogFatalf(ctx, "failed to connect database %v", err)
	}
	err = db.AutoMigrate(models...)
	if err != nil {
		runtime.LogFatalf(ctx, "failed to autoMigrate %v", err)
	}
	var v string
	result := db.Raw("SELECT sqlite_version()").Scan(&v)
	if result.Error != nil {
		runtime.LogFatalf(ctx, "failed to query sqlite version %v", result.Error)
	}
	runtime.LogInfof(ctx, "sqlite3 version: %s", v)
	model.ctx = ctx
}

type dbLoggerForTest struct{}

var dbLoggerForTestInstance = dbLoggerForTest{}

func (l dbLoggerForTest) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l dbLoggerForTest) Info(ctx context.Context, msg string, data ...interface{}) {
	// Discard logs
}

func (l dbLoggerForTest) Warn(ctx context.Context, msg string, data ...interface{}) {
	// Discard logs
}

func (l dbLoggerForTest) Error(ctx context.Context, msg string, data ...interface{}) {
	// Discard logs
}

func (l dbLoggerForTest) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// Discard logs
}
