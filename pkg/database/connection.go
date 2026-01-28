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
	*gorm.Config
}
