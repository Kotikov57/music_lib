//package routes представляет собой пакет для работы с маршрутами

package routes

import(
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"effect_mobile/models"
	"effect_mobile/db"
	"strconv"
	"strings"
)

//GetData получает все строки из базы данных
func GetData(c *gin.Context) {
	var data []models.Music 

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	result := db.Db.Limit(limit).Offset(offset).Find(&data)

	if result.Error != nil {
		c.JSON(400, gin.H{"error" : result.Error.Error()})
		return       
	}
	c.JSON(200, gin.H{
		"page" : page,
		"data" : data,
	})
}

//GetText получает текст конкретной песни
func GetText(c *gin.Context) {
	var text string
	songParam := c.Query("song")
	if songParam == "" {
		c.JSON(400, gin.H{"error" : "Название не может быть пустым"})
		return
	}
	result := db.Db.Where("song = ?", songParam).First(&text)
	if result.Error != nil {
		c.JSON(404, gin.H{"error" : "Песня не найдена"})
		return
	}
	verses := strings.Split(text, "\n\n")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1"))
	offset := (page - 1) * limit

	if offset >= len(verses) {
		c.JSON(200, gin.H{
			"page": page,
			"verse": "",
		})
		return
	}

	currentVerse := verses[offset]

	c.JSON(200, gin.H{
		"page" : page,
		"verse" : currentVerse,
	})
}

//DeleteData удаляет конкретную песню
func DeleteData(c *gin.Context) {
	songParam := c.Query("song")
	result := db.Db.Delete(&models.Music{}, songParam)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error" : "Песня не найдена"})
		} else {
			c.JSON(500, gin.H{"error" : result.Error.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message" : "Информация удалена"})
}

//PutData изменяет информация о конкретной песне
func PutData(c *gin.Context) {
	songParam := c.Query("song")
	var updatedData models.Music
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(400, gin.H{"error" : err.Error()})
		return
	}
	result := db.Db.Model(&updatedData).Where("song = ?", songParam).Updates(updatedData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error" : "Песня не найдена"})
		} else {
			c.JSON(500, gin.H{"error" : result.Error.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message" : "Информация изменена"})
}

func PostData(c *gin.Context) {
	var newData models.Music
	if err := c.ShouldBindJSON(&newData); err != nil {
		c.JSON(400, gin.H{"error" : err.Error()})
		return
	}
	result := db.Db.Create(&newData)
	if result.Error != nil {
		c.JSON(500, gin.H{"error" : result.Error.Error()})
		return
	}
	c.JSON(200, gin.H{"message" : "Информация добавлена"})
}