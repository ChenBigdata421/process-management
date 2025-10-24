package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConnection struct {
	db *gorm.DB
}

func NewPostgresConnection(dsn string) (*PostgresConnection, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresConnection{db: db}, nil
}

func (pc *PostgresConnection) GetDB() *gorm.DB {
	return pc.db
}

func (pc *PostgresConnection) Close() error {
	sqlDB, err := pc.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (pc *PostgresConnection) Migrate(models ...interface{}) error {
	return pc.db.AutoMigrate(models...)
}

func (pc *PostgresConnection) Health() error {
	sqlDB, err := pc.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func CreateDatabaseIfNotExists(adminDSN, dbName string) error {
	db, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database %s created successfully", dbName)
	} else {
		log.Printf("Database %s already exists", dbName)
	}

	return nil
}

