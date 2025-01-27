package main

import (
	"effect_mobile/db"
	_ "effect_mobile/docs"
	"effect_mobile/logger"
	"effect_mobile/routes"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

// @title Music API
// @version 1.0
// @description API для работы с музыкой
// @host localhost:8080
// @BasePath /

func main() {
	logger.InitLogger()
	logger.Log.Info("[INFO] Приложение запущено")
	router := gin.Default()
	db.ConnectDatabase()
	defer db.CloseDatabase()
	db.RunMigrations()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/info", routes.GetData)
	router.GET("/texts", routes.GetText)
	router.DELETE("/info", routes.DeleteData)
	router.PUT("/info", routes.PutData)
	router.PUT("/info/param", routes.PutParam)
	router.POST("/info", routes.PostData)

	router.Run(":8080")
}
