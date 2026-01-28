package database

type Repository interface {
	Initialize(connection Connection) error
}
