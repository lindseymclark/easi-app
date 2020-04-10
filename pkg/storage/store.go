package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store performs database operations for EASi
type Store struct {
	DB     *sqlx.DB // temporarily export until SystemIntakesHandler doesn't take db
	logger *zap.Logger
}

// DBConfig holds the configurations for a database connection
type DBConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

// NewStore is a constructor for a store
func NewStore(
	logger *zap.Logger,
	config DBConfig,
) (*Store, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
	)
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &Store{DB: db, logger: logger}, nil
}