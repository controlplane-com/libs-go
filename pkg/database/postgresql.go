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

func postgresqlDb(c *ConnectionConfiguration) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.Host, c.User, c.Password, c.Database, c.Port)
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
