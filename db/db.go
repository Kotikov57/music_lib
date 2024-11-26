//package db представляет собой пакет для работы с PostgreSQ

package db

import (
	"database/sql"
	"effect_mobile/envutils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

var Db *sql.DB

// ConnectDatabase подлючает к базе данных
func ConnectDatabase() {
	log.Println("[DEBUG] Вход в функцию ConnectDatabase")
	dsn := envutils.GetDatabaseURL()
	var err error
	Db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения в базе данных:", err)
		return
	}
	if err := Db.Ping(); err != nil {
		log.Fatal("Ошибка проверки соединения: ", err)
	}
	log.Println("[INFO] База данных подключена")
}

// CloseDatabase закрывает базу данных
func CloseDatabase() {
	log.Println("[DEBUG] Вход в функцию CloseDatabase")
	Db.Close()
	log.Println("[INFO] База данных закрыта")
}

// RunMigrations запускает миграции
func RunMigrations() {
	log.Println("[DEBUG] Вход в функцию CloseDatabase")
	driver, err := postgres.WithInstance(Db, &postgres.Config{})
	if err != nil {
		log.Fatal("Ошибка создания драйвера миграции:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		envutils.GetMigrationsDir(),
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("Ошибка инициализации миграции:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Ошибка выполнения миграции:", err)
	}
	log.Println("[INFO] Миграция выполнена")
}
