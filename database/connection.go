package database

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func ConnectToDB(dbUser, dbPassword, dbName string) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable", 
        dbUser, dbPassword, dbName)

    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}