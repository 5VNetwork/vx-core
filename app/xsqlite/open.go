// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package xsqlite

import (
	"fmt"
	"runtime"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// OpenSharedFile opens a SQLite file that may also be used by another SQLite
// library in the same process (e.g. Dart/Drift on Android).
//
// On Android it pins a single never-recycled connection so pool close() churn
// does not trip the POSIX advisory-lock bug with a second SQLite copy.
func OpenSharedFile(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path+"?_foreign_keys=on&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	if err := configureSharedFile(db); err != nil {
		_ = closeGorm(db)
		return nil, err
	}
	return db, nil
}

func configureSharedFile(db *gorm.DB) error {
	if err := db.Exec("PRAGMA journal_mode = WAL").Error; err != nil {
		log.Warn().Err(err).Msg("failed to enable WAL journal mode")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if runtime.GOOS == "android" {
		// One long-lived connection: avoid open/close churn that drops locks
		// held by the Dart-side SQLite library in the same process.
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(0)
		sqlDB.SetConnMaxIdleTime(0)
	}
	return nil
}

func closeGorm(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
