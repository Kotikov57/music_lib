// utils Пакет предназначенный для работы с env-файлом
package envutils

import (
	"effect_mobile/logger"
	"github.com/joho/godotenv"
	"os"
)

// init подключает env-файл
func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatal("[FATAL] Не удалось подключить env-файл")
	}

}

// GetEnvVariable получает значение переменной окружения из env-файла
func GetEnvVariable(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		logger.Log.Error("[ERROR] Перменная окружения не задана", key)
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

//GetOutAPI получается из env-файла адрес внешнего API
func GetOutAPI() string {
	return GetEnvVariable("OUT_API_URL")
}
