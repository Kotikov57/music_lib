package main

import (
	"effect_mobile/db"
	"effect_mobile/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	db.ConnectDatabase()

	router.GET("/info", routes.GetData)
	router.GET("/info", routes.GetText)
	router.DELETE("/info", routes.DeleteData)
	router.PUT("/info", routes.PutData)
	router.POST("/info", routes.PostData)

	router.Run(":8080")
}