package database

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresqlConnection struct {
	ReadWriteConfiguration *ConnectionConfiguration
	ReadOnlyConfiguration  *ConnectionConfiguration
	db                     *gorm.DB
	dbRo                   *gorm.DB
}

func NewPostgresqlConnection(readWriteConfiguration *ConnectionConfiguration, readOnlyConfiguration *ConnectionConfiguration) (Connection, error) {
	if readWriteConfiguration == nil {
		return nil, errors.New("received a nil readWriteConfiguration. A postgresql repository requires at least read/write database connection")
	}
	return &PostgresqlConnection{
		ReadWriteConfiguration: readWriteConfiguration,
		ReadOnlyConfiguration:  readOnlyConfiguration,
	}, nil
}

func (p *PostgresqlConnection) Initialize() error {
	if p.db != nil {
		return nil
	}
	db, err := postgresqlDb(p.ReadWriteConfiguration)
	if err != nil {
		return err
	}
	p.db = db

	var dbRo *gorm.DB
	if p.ReadOnlyConfiguration == nil {
		p.dbRo = db
		return nil
	}

	dbRo, err = postgresqlDb(p.ReadOnlyConfiguration)
	if err != nil {
		return err
	}
	p.dbRo = dbRo
	return nil
}

func (p *PostgresqlConnection) Db() *gorm.DB {
	return p.db
}

func (p *PostgresqlConnection) DbRo() *gorm.DB {
	return p.dbRo
}

// buildDSN assembles the libpq keyword/value connection string, including the
// optional server-side statement_timeout (ports connection.ts afterCreate `SET
// statement_timeout`). Extracted for unit testing.
func buildDSN(c *ConnectionConfiguration) string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.Host, c.User, c.Password, c.Database, c.Port)
	if c.StatementTimeout > 0 {
		// `options='-c statement_timeout=<ms>'` is a startup parameter applied to every
		// physical connection — the driver-agnostic equivalent of a per-connection SET.
		dsn += fmt.Sprintf(" options='-c statement_timeout=%d'", c.StatementTimeout.Milliseconds())
	}
	return dsn
}

func postgresqlDb(c *ConnectionConfiguration) (*gorm.DB, error) {
	dsn := buildDSN(c)
	gormDb, err := gorm.Open(postgres.Open(dsn), c.Config)
	if err != nil {
		return nil, err
	}
	db, err := gormDb.DB()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(c.MaxIdleConnections)
	db.SetConnMaxIdleTime(c.MaxIdleConnectionLifetime)
	db.SetMaxOpenConns(c.MaxOpenConnections)
	db.SetConnMaxLifetime(c.MaxConnectionLifetime)
	return gormDb, nil
}
