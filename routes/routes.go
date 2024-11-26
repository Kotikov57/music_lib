//package routes представляет собой пакет для работы с маршрутами

package routes

import (
	"database/sql"
	"effect_mobile/db"
	_ "effect_mobile/docs"
	"effect_mobile/models"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GetData получает все строки из базы данных
// @Summary Получить данные о песнях
// @Description Возвращает все данные о всех песнях
// @Tags Music
// @Success 200 {object} models.Music
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [get]
func GetData(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию GetData")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	log.Printf("[DEBUG] page = %d | limit = %d | offset = %d", page, limit, offset)
	query := "SELECT * FROM music LIMIT $1 OFFSET $2"
	rows, err := db.Db.Query(query, limit, offset)

	if err != nil {
		log.Println("[ERROR] Ошибка выполнения запроса:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}
	defer rows.Close()

	var data []models.Music

	for rows.Next() {
		var m models.Music
		err := rows.Scan(&m.Main.Group, &m.Main.Song, &m.Details.ReleaseDate, &m.Details.Text, &m.Details.Link)
		if err != nil {
			log.Println("[ERROR] Ошибка сканирования строки:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		log.Printf("[DEBUG] Group = %s | Song = %s | ReleaseDate = %s | Text = %s | Link = %s", m.Main.Group, m.Main.Song, m.Details.ReleaseDate, m.Details.Text, m.Details.Link)
		data = append(data, m)
	}

	if err = rows.Err(); err != nil {
		log.Println("[ERROR] Ошибка обработки строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка чтения строк"})
		return
	}

	c.JSON(200, gin.H{
		"page": page,
		"data": data,
	})
	log.Println("[INFO] Запрос успешно выполнен")
}

// GetText получает из базы данных текст песни по параметрам song и group
// @Summary Получить текст песни
// @Description Возвращает текст песни по группе и названию
// @Tags Music
// @Param group query string true "Группа"
// @Param song query string true "Песня"
// @Success 200 {object} string
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /texts [get]
func GetText(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию GetText")
	groupParam := c.Query("group")
	songParam := c.Query("song")
	log.Printf("[DEBUG] groupParam = %s", groupParam)
	log.Printf("[DEBUG] songParam = %s", songParam)
	if groupParam == "" {
		log.Println("[ERROR] Название группы не может быть пустым")
		c.JSON(400, gin.H{"error": "Название группы не может быть пустым"})
		return
	}
	if songParam == "" {
		log.Println("[ERROR] Название песни не может быть пустым")
		c.JSON(400, gin.H{"error": "Название песни не может быть пустым"})
		return
	}

	var text string
	err := db.Db.QueryRow(`SELECT text FROM music WHERE "group" = $1 AND song = $2`, groupParam, songParam).Scan(&text)
	if err == sql.ErrNoRows {
		log.Println("[ERROR] Песня не найдена")
		c.JSON(404, gin.H{"error": "Песня не найдена"})
		return
	} else if err != nil {
		log.Println("[ERROR] Ошибка запроса к базе данных")
		c.JSON(500, gin.H{"error": "Ошибка запроса к базе данных"})
		return
	}
	log.Printf("[DEBUG] text = %s", text)
	verses := strings.Split(text, "\n\n")
	for i, v := range verses {
		log.Printf("[DEBUG] verse %d: %s", i+1, v)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1"))
	offset := (page - 1) * limit
	log.Printf("[DEBUG] page = %d | limit = %d | offset = %d", page, limit, offset)

	if offset >= len(verses) {
		c.JSON(200, gin.H{
			"page":  page,
			"verse": "",
		})
		return
	}

	currentVerse := verses[offset]
	log.Printf("[DEBUG] currentVerse = %s", currentVerse)

	c.JSON(200, gin.H{
		"page":  page,
		"verse": currentVerse,
	})
	log.Println("[INFO] Запрос успешно выполнен")
}

// DeleteData удаляет из базы данных информацию по параметру song
// @Summary Удалить данные песни
// @Description удаляет данные песни по группе и названию
// @Tags Music
// @Param group query string true "Группа"
// @Param song query string true "Песня"
// @Success 200 {object} string
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [delete]
func DeleteData(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию DeleteData")
	groupParam := c.Query("group")
	log.Printf("[DEBUG] groupParam = %s", groupParam)
	songParam := c.Query("song")
	log.Printf("[DEBUG] songParam = %s", songParam)

	result, err := db.Db.Exec(`DELETE FROM music WHERE "group = $1 AND song = $2`, groupParam, songParam)
	if err != nil {
		log.Println("[ERROR] Ошибка при удалении записи:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("[ERROR] Ошибка получения затронутых строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
		return
	}
	if rowsAffected == 0 {
		log.Println("[ERROR] Запись не найдена:", err)
		c.JSON(404, gin.H{"error": "Запись не найдена"})
		return
	}

	c.JSON(200, gin.H{"message": "Информация удалена"})
	log.Println("[INFO] Запрос успешно выполнен")
}

// PutData изменяет информация о конкретной песне в базе данных
// @Summary Изменить данные песни
// @Description Изменяет данные песни по группе и названию
// @Tags Music
// @Param group query string true "Группа"
// @Param song query string true "Песня"
// @Success 200 {object} string
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [put]
func PutData(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию PutData")
	groupParam := c.Query("group")
	log.Printf("[DEBUG] groupParam = %s", groupParam)
	songParam := c.Query("song")
	log.Printf("[DEBUG] songParam = %s", songParam)

	var updatedData models.Music
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		log.Println("[ERROR] Неверный формат данных:", err)
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}

	query := `UPDATE music SET "group" = $1, song = $2, releaseDate = $3, text = $4, link = $5 WHERE "group" = $6 AND song = $7`
	result, err := db.Db.Exec(query, updatedData.Main.Group, updatedData.Main.Song, updatedData.Details.ReleaseDate, updatedData.Details.Text, updatedData.Details.Link, groupParam, songParam)
	if err != nil {
		log.Println("[ERROR] Ошибка выполнения запроса:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("[ERROR] Ошибка получения количества строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
		return
	}
	if rowsAffected == 0 {
		log.Println("[ERROR] Запись не найдена:", err)
		c.JSON(404, gin.H{"error": "Запись не найдена"})
		return
	}

	c.JSON(200, gin.H{"message": "Информация обновлена"})
	log.Println("[INFO] Запрос успешно выполнен")
}

// PostData добавляет информацию о песне в базу данных
// @Summary Добавить данные песни
// @Description Добавляет данные песни и делает запрос в внешний АПИ для получения дополнительных данных
// @Tags Music
// @Accept json
// @Param music body models.Song true "Группа и название"
// @Success 200 {object} string
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [post]
func PostData(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию PutData")
	var newSong models.Song
	if err := c.ShouldBindJSON(&newSong); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	err := db.Db.QueryRow(`SELECT EXISTS (SELECT  1 FROM music WHERE "group" = $1 AND song = $2 LIMIT 1)`, newSong.Group, newSong.Song).Scan(&exists)
	if err != nil {
		log.Println("[ERROR] Ошибка при проверке существования песни:", err)
		c.JSON(500, gin.H{"error": "Ошибка при проверке существования песни: " + err.Error()})
		return
	}
	if exists {
		log.Println("[ERROR] Песня уже существует")
		c.JSON(400, gin.H{"error": "Песня уже существует"})
		return
	}

	var newData models.Music
	newData.Main = newSong
	songDetails := getSongDetailsFromAPI(newSong.Group, newSong.Song)
	if songDetails == nil {
		log.Println("[ERROR] Ошибка при получении данных из АПИ")
		c.JSON(500, gin.H{"error": "Ошибка при получении данных из АПИ"})
		return
	}

	newData.Details.ReleaseDate = songDetails.ReleaseDate
	newData.Details.Text = songDetails.Text
	newData.Details.Link = songDetails.Link
	log.Printf("[DEBUG] newData.Group = %s | newData.Song = %s | newData.ReleaseDate = %s | newData.Text = %s | newData.Link = %s", newData.Main.Group, newData.Main.Song, newData.Details.ReleaseDate, newData.Details.Text, newData.Details.Link)

	query := `INSERT INTO music ("group", song, releaseDate, text, link) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.Db.Exec(query, &newData.Main.Group, &newData.Main.Song, &newData.Details.ReleaseDate, &newData.Details.Text, newData.Details.Link)

	if err != nil {
		log.Println("[ERROR] Ошибка добавления данных:", err)
		c.JSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"message": "Информация добавлена"})
	log.Println("[INFO] Запрос успешно выполнен")
}

// getSongDetailsFromAPI делает запрос в внешний АПИ для получения дополнительных данных (releaseDate, text, link)
func getSongDetailsFromAPI(group string, song string) *models.SongDetail {
	log.Println("[DEBUG] Вход в функцию getSongDetailsFromAPI")
	url := fmt.Sprintf("http:/some-api.com/info?group=%s&song=%s", group, song)
	log.Printf("[DEBUG] url = %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("[ERROR] Ошибка при получении данных песни: ", err)
		return nil
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] resp.StatusCode = %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		log.Println("[ERROR] АПИ вернул не 200: ", resp.StatusCode)
		return nil
	}

	var songDetails models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		log.Println("[ERROR] Ошибка при декодировании ответа АПИ: ", err)
		return nil
	}
	log.Printf("[DEBUG] songDetails = %v", songDetails)
	log.Println("[INFO] Запрос к АПИ успешно сделан")
	return &songDetails
}
