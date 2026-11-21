package database

import (
	"gorm.io/gorm"
	"time"
)

type Connection interface {
	Initialize() error
	Db() *gorm.DB
	DbRo() *gorm.DB
}

type ConnectionConfiguration struct {
	Host                      string
	Port                      string
	User                      string
	Password                  string
	Database                  string
	MaxIdleConnections        int
	MaxOpenConnections        int
	MaxIdleConnectionLifetime time.Duration
	MaxConnectionLifetime     time.Duration
	// StatementTimeout, when > 0, caps every query server-side by adding
	// `options='-c statement_timeout=<ms>'` to the DSN — a startup parameter applied
	// to every physical connection (equivalent to a per-connection SET). Opt-in so a
	// service with legitimately long queries is unaffected; 0 leaves it unset.
	StatementTimeout time.Duration
	*gorm.Config
}
