// utils Пакет предназначенный для работы с env-файлом
package envutils

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// init подключает env-файл
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] Не удалось подключить env-файл")
	}

}

// GetEnvVariable получает значение переменной окружения из env-файла
func GetEnvVariable(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("[ERROR] Перменная окружения %s не задана", key)
	}
	return value
}

// GetDatabaseURL получает из env-файла ссылку для подключения в базе данных
func GetDatabaseURL() string {
	return GetEnvVariable("DATABASE_URL")
}

// GetMigrationsDir получает из env-файла директорию с миграциями
func GetMigrationsDir() string {
	return GetEnvVariable("MIGRATIONS_DIR")
}
