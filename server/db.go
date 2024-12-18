package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

var (
	// Do not use this in production code. Use DB() instead.
	_db *gorm.DB

	errCodes = map[string]error{
		"23505": gorm.ErrDuplicatedKey,
		"23503": gorm.ErrForeignKeyViolated,
		"42703": gorm.ErrInvalidField,
	}
)

func DB() *gorm.DB { return _db.WithContext(ctx()) }

func RawDB() *gorm.DB { return _db }

func ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	return ctx
}

func OpenDB() error {
	if _db != nil {
		return nil
	}
	ll := gorm_logger.Warn | gorm_logger.Error
	if IsEnvLocal() {
		ll = ll | gorm_logger.Info
	}
	dbLogger := gorm_logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		gorm_logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  ll,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	dsn := os.Getenv("POSTGRESQL_URL")
	if dsn == "" {
		return errors.New("POSTGRESQL_URL env variable not set")
	}
	postgres, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: dbLogger, CreateBatchSize: 100})
	if err != nil {
		return errors.New("Failed to connect to Postgres: " + err.Error())
	}
	_db = postgres

	return nil
}

func CloseDB() error {
	db, _ := _db.DB()
	return db.Close()
}

func Ping() error { return _db.Exec("SELECT 1").Error }

/// Error Checking

// TranslateError translates the error to native gorm errors.
// only works for pgx PgError types.
func TransPGErrors(err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if translatedErr, found := errCodes[pgErr.Code]; found {
			return translatedErr
		}
	}
	return err
}
