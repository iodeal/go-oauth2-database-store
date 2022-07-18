package dbstore

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

// StoreItem data item
type StoreItem struct {
	ID        int64  `db:"id"`
	ExpiredAt int64  `db:"expired_at"`
	Code      string `db:"code"`
	Access    string `db:"access"`
	Refresh   string `db:"refresh"`
	Data      string `db:"data"`
}

// NewConfig create database configuration instance
func NewConfig(driverName, dsn string) *Config {
	return &Config{
		DriverName:   driverName,
		DSN:          dsn,
		MaxLifetime:  time.Hour * 2,
		MaxOpenConns: 50,
		MaxIdleConns: 25,
	}
}

// Config database configuration
type Config struct {
	DriverName   string
	DSN          string
	MaxLifetime  time.Duration
	MaxOpenConns int
	MaxIdleConns int
}

// NewDefaultStore create database store instance
func NewDefaultStore(config *Config) *Store {
	return NewStore(config, "", 0)
}

// NewStore create database store instance,
// config database configuration,
// tableName table name (default oauth2_token),
// GC time interval (in seconds, default 600)
func NewStore(config *Config, tableName string, gcInterval int) *Store {
	db, err := sqlx.Open(config.DriverName, config.DSN)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.MaxLifetime)

	return NewStoreWithDB(db, tableName, gcInterval)
}

// NewStoreWithDB create database store instance,
// db sql.DB,
// tableName table name (default oauth2_token),
// GC time interval (in seconds, default 600)
func NewStoreWithDB(db *sqlx.DB, tableName string, gcInterval int) *Store {
	// Init store with options
	store := NewStoreWithOpts(db,
		WithTableName(tableName),
		WithGCTimeInterval(gcInterval),
	)

	go store.gc()
	return store
}

// NewStoreWithOpts create database store instance with apply custom input,
// db sql.DB,
// tableName table name (default oauth2_token),
// GC time interval (in seconds, default 600)
func NewStoreWithOpts(db *sqlx.DB, opts ...Option) *Store {
	// Init store with default value
	store := &Store{
		db:        db,
		tableName: "oauth2_token",
		stdout:    os.Stderr,
		ticker:    time.NewTicker(time.Second * time.Duration(600)),
	}

	// Apply with optional function
	for _, opt := range opts {
		opt.apply(store)
	}
	var n int
	err := store.db.QueryRowx(fmt.Sprintf("select 1 from %s limit 1", store.tableName)).Scan(&n)
	if err != nil && err != sql.ErrNoRows {
		_, err = store.db.Exec(fmt.Sprintf(`create table if not exists %s (
			id bigint not null primary key,
			expired_at bigint,
			code varchar(255),
			access varchar(255),
			refresh varchar(255),
			data varchar(2048)
		)`, store.tableName))
		if err != nil {
			panic(err)
		}
		store.db.Exec(fmt.Sprintf("create index idx_%s_code on %s(code))", store.tableName, store.tableName))
		store.db.Exec(fmt.Sprintf("create index idx_%s_access on %s(access))", store.tableName, store.tableName))
		store.db.Exec(fmt.Sprintf("create index idx_%s_refresh on %s(refresh))", store.tableName, store.tableName))
		store.db.Exec(fmt.Sprintf("create index idx_%s_expired_at on %s(expired_at))", store.tableName, store.tableName))
	}
	go store.gc()
	return store
}
