//package db представляет собой пакет для работы с PostgreSQ

package db

import (
	"database/sql"
	"log"
	"github.com/jackc/pgx/v5/stdlib"
)

var Db *gorm.DB
 
// ConnectDatabase подлючает к базе данных
func ConnectDatabase() {
	dsn := "host=localhost user=postgres password=fkla5283 dbname= imarket_db port=5432 sslmode=disable TimeZone=UTC"
	var err error
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return
	}
}