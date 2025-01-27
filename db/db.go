//package db представляет собой пакет для работы с PostgreSQ

package db

import (
	"database/sql"
	"effect_mobile/envutils"
	"effect_mobile/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var Db *sql.DB

// ConnectDatabase подлючает к базе данных
func ConnectDatabase() {
	logger.Log.Debug("[DEBUG] Вход в функцию ConnectDatabase")
	dsn := envutils.GetDatabaseURL()
	var err error
	Db, err = sql.Open("pgx", dsn)
	if err != nil {
		logger.Log.Fatal("Ошибка подключения в базе данных:", err)
		return
	}
	if err := Db.Ping(); err != nil {
		logger.Log.Fatal("Ошибка проверки соединения: ", err)
	}
	logger.Log.Info("[INFO] База данных подключена")
}

// CloseDatabase закрывает базу данных
func CloseDatabase() {
	logger.Log.Debug("[DEBUG] Вход в функцию CloseDatabase")
	Db.Close()
	logger.Log.Info("[INFO] База данных закрыта")
}

// RunMigrations запускает миграции
func RunMigrations() {
	logger.Log.Debug("[DEBUG] Вход в функцию CloseDatabase")
	driver, err := postgres.WithInstance(Db, &postgres.Config{})
	if err != nil {
		logger.Log.Fatal("Ошибка создания драйвера миграции:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		envutils.GetMigrationsDir(),
		"postgres",
		driver,
	)
	if err != nil {
		logger.Log.Fatal("Ошибка инициализации миграции:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Fatal("Ошибка выполнения миграции:", err)
	}
	logger.Log.Info("[INFO] Миграция выполнена")
}
