//package routes представляет собой пакет для работы с маршрутами

package routes

import (
	"database/sql"
	"effect_mobile/db"
	"effect_mobile/logger"
	"time"
	//	"effect_mobile/docs"
	"effect_mobile/envutils"
	"effect_mobile/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetData получает все строки из базы данных
// @Summary Получить данные о песнях
// @Description Возвращает все данные о всех песнях
// @Tags Music
// @Success 200 {object} models.Music
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [get]
func GetData(c *gin.Context) {
	logger.Log.Debug("[DEBUG] Вход в функцию GetData")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	groupName := c.Query("group")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	keyword := c.Query("keyword")
	debugParam := fmt.Sprintf("[DEBUG] page = %d | limit = %d | offset = %d", page, limit, offset)
	logger.Log.Debug(debugParam)

	query := "SELECT groups.group_name AS group, songs.song_name AS song, release_date, text, link " +
		"FROM details " +
		"JOIN songs ON details.song_id = songs.song_id " +
		"JOIN groups ON songs.group_id = groups.group_id " +
		"WHERE ($1 = '' OR groups.group_name = $1) " +
		"AND ($2 = '' OR details.release_date >= TO_DATE($2, 'DD-MM-YYYY')) AND ($3 = '' OR details.release_date <= TO_DATE($3, 'DD-MM-YYYY')) " +
		"AND ($4 = '' OR details.text ILIKE '%' || $4 || '%') " +
		"ORDER BY details.release_date DESC " +
		"LIMIT $5 OFFSET $6"

	rows, err := db.Db.Query(query, groupName, startDate, endDate, keyword, limit, offset)

	if err != nil {
		logger.Log.Error("[ERROR] Ошибка выполнения запроса:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}
	defer rows.Close()

	var data []models.MusicRequest

	for rows.Next() {
		var req models.MusicRequest
		var m models.Music
		err := rows.Scan(&m.Main.Group, &m.Main.Song, &m.Details.ReleaseDate, &m.Details.Text, &m.Details.Link)
		if err != nil {
			logger.Log.Error("[ERROR] Ошибка сканирования строки:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		req.Main.Group = m.Main.Group
		req.Main.Song = m.Main.Song
		req.Details.ReleaseDate = m.Details.ReleaseDate.Format("02-01-2006")
		req.Details.Text = m.Details.Text
		req.Details.Link = m.Details.Link
		debugString := fmt.Sprintf("[DEBUG] Group = %s | Song = %s | ReleaseDate = %s | Text = %s | Link = %s",
			req.Main.Group, req.Main.Song, req.Details.ReleaseDate, req.Details.Text, req.Details.Link)
		logger.Log.Debug(debugString)

		data = append(data, req)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error("[ERROR] Ошибка обработки строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка чтения строк"})
		return
	}

	c.JSON(200, gin.H{
		"page": page,
		"data": data,
	})
	logger.Log.Info("[INFO] Запрос успешно выполнен")
}

// GetText получает из базы данных текст песни по параметру song
// @Summary Получить текст песни
// @Description Возвращает текст песни по её ID
// @Tags Music
// @Param song query string true "ID песни"
// @Success 200 {object} string
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /texts [get]
func GetText(c *gin.Context) {
	logger.Log.Debug("[DEBUG] Вход в функцию GetText")
	songIdParam := c.Query("song")
	DebugParam := fmt.Sprintf("[DEBUG] songParam = %s", songIdParam)
	logger.Log.Debug(DebugParam)

	var text string
	err := db.Db.QueryRow(`SELECT text FROM details `+
		`WHERE song_id = $1`, songIdParam).Scan(&text)
	if err == sql.ErrNoRows {
		logger.Log.Error("[ERROR] Песня не найдена")
		c.JSON(404, gin.H{"error": "Песня не найдена"})
		return
	} else if err != nil {
		logger.Log.Error("[ERROR] Ошибка запроса к базе данных:", err)
		c.JSON(500, gin.H{"error": "Ошибка запроса к базе данных"})
		return
	}
	DebugParam = fmt.Sprintf("[DEBUG] text = %s", text)
	logger.Log.Debug(DebugParam)
	verses := strings.Split(text, "\n\n")
	for i, v := range verses {
		DebugParam = fmt.Sprintf("[DEBUG] verse %d: %s", i+1, v)
		logger.Log.Debug(DebugParam)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1"))
	offset := (page - 1) * limit
	DebugParam = fmt.Sprintf("[DEBUG] page = %d | limit = %d | offset = %d", page, limit, offset)
	logger.Log.Debug(DebugParam)

	if offset >= len(verses) {
		c.JSON(200, gin.H{
			"page":  page,
			"verse": "",
		})
		return
	}

	currentVerse := verses[offset]
	DebugParam = fmt.Sprintf("[DEBUG] currentVerse = %s", currentVerse)
	logger.Log.Debug(DebugParam)

	c.JSON(200, gin.H{
		"page":  page,
		"verse": currentVerse,
	})
	logger.Log.Info("[INFO] Запрос успешно выполнен")
}

// DeleteData удаляет из базы данных информацию по параметру song
// @Summary Удалить данные песни
// @Description удаляет данные песни по её ID
// @Tags Music
// @Param song query string true "ID песни"
// @Success 200 {object} string
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [delete]
func DeleteData(c *gin.Context) {
	logger.Log.Debug("[DEBUG] Вход в функцию DeleteData")
	songIdParam := c.Query("song")
	DebugParam := fmt.Sprintf("[DEBUG] songParam = %s", songIdParam)
	logger.Log.Debug(DebugParam)

	tx, err := db.Db.Begin()
	if err != nil {
		logger.Log.Error("[ERROR] Не удалось начать транзакцию")
		c.JSON(500, gin.H{"error": "Не получилось начать транзакцию"})
	}
	defer tx.Rollback()
	result1, err := tx.Exec(`DELETE FROM details WHERE song_id = $1`, songIdParam)
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка при удалении записи из details:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}
	result2, err := tx.Exec(`DELETE FROM songs WHERE song_id = $1`, songIdParam)
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка при удалении записи из songs:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("[ERROR] Не удалось закоммитить транзакцию")
		c.JSON(500, gin.H{"error": "Не удалось закоммитить транзакцию"})
		return
	}
	rowsAffected1, err := result1.RowsAffected()
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка получения затронутых строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
		return
	}
	rowsAffected2, err := result2.RowsAffected()
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка получения затронутых строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
		return
	}
	if rowsAffected1 == 0 || rowsAffected2 == 0 {
		logger.Log.Error("[ERROR] Запись не найдена:", err)
		c.JSON(404, gin.H{"error": "Запись не найдена"})
		return
	}

	c.JSON(200, gin.H{"message": "Информация удалена"})
	logger.Log.Info("[INFO] Запрос успешно выполнен")
}

// PutData изменяет информация о конкретной песне в базе данных
// @Summary Изменить данные песни
// @Description Изменяет данные песни по её ID
// @Tags Music
// @Param song query string true "ID песни"
// @Success 200 {object} string
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /info [put]
func PutData(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию PutData")
	songIdParam := c.Query("song")
	DebugParam := fmt.Sprintf("[DEBUG] songParam = %s", songIdParam)
	logger.Log.Debug(DebugParam)

	var updatedDataRequest struct {
		Song_name    string `json:"song_name"`
		Release_date string `json:"release_date"`
		Text         string `json:"text"`
		Link         string `json:"link"`
	}
	if err := c.ShouldBindJSON(&updatedDataRequest); err != nil {
		logger.Log.Error("[ERROR] Неверный формат данных:", err)
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}
	tx, err := db.Db.Begin()
	if err != nil {
		logger.Log.Error("[ERROR] Не удалось начать транзакцию")
		c.JSON(500, gin.H{"error": "Не получилось начать транзакцию"})
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE songs SET song_name = $1 WHERE song_id = $2", updatedDataRequest.Song_name, songIdParam)
	if err != nil {
		logger.Log.Error("[ERROR] Не удалось обновить название песни")
		c.JSON(500, gin.H{"error": "Не удалось обновить название песни"})
		return
	}

	_, err = tx.Exec("UPDATE details SET release_date = TO_DATE($1,'DD-MM-YYYY'), text = $2, link = $3 WHERE song_id = $4",
		updatedDataRequest.Release_date, updatedDataRequest.Text, updatedDataRequest.Link, songIdParam)

	if err != nil {
		logger.Log.Error("[ERROR] Не удалось обновить данные песни")
		c.JSON(500, gin.H{"error": "Не удалось обновить данные песни"})
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("[ERROR] Не удалось закоммитить транзакцию")
		c.JSON(500, gin.H{"error": "Не удалось закоммитить транзакцию"})
		return
	}

	c.JSON(200, gin.H{"message": "Информация обновлена"})
	logger.Log.Info("[INFO] Запрос успешно выполнен")
}

// PutParam изменяет конкретную информацию о конкретной песне в базе данных
// @Summary Изменить данные песни
// @Description Изменяет данные песни по её ID
// @Tags Music
// @Param song query string true "ID песни"
// @Success 200 {object} string
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /info/param [put]
func PutParam(c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию PutData")
	songIdParam := c.Query("song")
	DebugParam := fmt.Sprintf("[DEBUG] songParam = %s", songIdParam)
	logger.Log.Debug(DebugParam)
	param := c.Query("param")

	switch param {
	case "name":
		var updatedSong struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&updatedSong); err != nil {
			logger.Log.Error("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE songs SET song_name = $1 WHERE song_id = $2", updatedSong.Name, songIdParam)
		if err != nil {
			logger.Log.Error("[ERROR] Не удалось обновить название песни")
			c.JSON(500, gin.H{"error": "Не удалось обновить название песни"})
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Log.Error("[ERROR] Ошибка получения количества строк:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		if rowsAffected == 0 {
			logger.Log.Error("[ERROR] Запись не найдена:", err)
			c.JSON(404, gin.H{"error": "Запись не найдена"})
			return
		}
	case "release_date":
		var updatedDate struct {
			Date string `json:"date"`
		}
		if err := c.ShouldBindJSON(&updatedDate); err != nil {
			logger.Log.Error("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE details SET release_date = TO_DATE($1, 'DD-MM-YYYY') WHERE song_id = $2", updatedDate.Date, songIdParam)
		if err != nil {
			logger.Log.Error("[ERROR] Не удалось обновить дату релиза песни:", err)
			c.JSON(500, gin.H{"error": "Не удалось обновить дату релиза песни"})
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Log.Error("[ERROR] Ошибка получения количества строк:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		if rowsAffected == 0 {
			logger.Log.Error("[ERROR] Запись не найдена:", err)
			c.JSON(404, gin.H{"error": "Запись не найдена"})
			return
		}
	case "text":
		var updatedText struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&updatedText); err != nil {
			logger.Log.Error("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE details SET text = $1 WHERE song_id = $2", updatedText.Text, songIdParam)
		if err != nil {
			logger.Log.Error("[ERROR] Не удалось обновить текст песни:", err)
			c.JSON(500, gin.H{"error": "Не удалось обновить текст песни"})
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Log.Error("[ERROR] Ошибка получения количества строк:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		if rowsAffected == 0 {
			logger.Log.Error("[ERROR] Запись не найдена:", err)
			c.JSON(404, gin.H{"error": "Запись не найдена"})
			return
		}
	case "link":
		var updatedLink struct {
			Link string `json:"link"`
		}
		if err := c.ShouldBindJSON(&updatedLink); err != nil {
			logger.Log.Error("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE details SET link = $1 WHERE song_id = $2", updatedLink.Link, songIdParam)
		if err != nil {
			logger.Log.Error("[ERROR]Не удалось обновить ссылку на песню:", err)
			c.JSON(500, gin.H{"error": "Не удалось обновить ссылку на песню"})
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Log.Error("[ERROR] Ошибка получения количества строк:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		if rowsAffected == 0 {
			logger.Log.Error("[ERROR] Запись не найдена:", err)
			c.JSON(404, gin.H{"error": "Запись не найдена"})
			return
		}
	default:
		logger.Log.Error("[ERROR] Некорректный параметр")
		c.JSON(400, gin.H{"error": "Некорректный параметр"})
	}

	c.JSON(200, gin.H{"message": "Информация обновлена"})
	logger.Log.Info("[INFO] Запрос успешно выполнен")
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
	logger.Log.Debug("[DEBUG] Вход в функцию PostData")
	var newSong models.Song
	if err := c.ShouldBindJSON(&newSong); err != nil {
		logger.Log.Error("[ERROR] Некорректный формат данных")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	_, err := db.Db.Exec(`
		INSERT INTO groups (group_name)
		VALUES ($1)
		ON CONFLICT (group_name) DO NOTHING;
	`, newSong.Group)
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка при добавлении группы")
		c.JSON(500, gin.H{"error": "Ошибка при добавлении группы"})
		return
	}
	_, err = db.Db.Exec(`
		INSERT INTO songs (group_id, song_name)
		VALUES (
			(SELECT group_id FROM groups WHERE group_name = $1),
			$2
		)
		ON CONFLICT (group_id, song_name) DO NOTHING;
	`, newSong.Group, newSong.Song)
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка при добавлении песни")
		c.JSON(500, gin.H{"error": "Ошибка при добавлении песни"})
		return
	}

	var newData models.Music
	newData.Main = newSong
	songDetails := getSongDetailsFromAPI(newSong.Group, newSong.Song)
	if songDetails == nil {
		logger.Log.Error("[ERROR] Ошибка при получении данных из АПИ")
		c.JSON(500, gin.H{"error": "Ошибка при получении данных из АПИ"})
		return
	}
	date, err := time.Parse("02-01-2006", songDetails.ReleaseDate)
	if err != nil {
		logger.Log.Error("[ERROR] Неверный формат времени:", err)
		c.JSON(400, gin.H{"error": "Неверный формат времени"})
		return
	}
	newData.Details.ReleaseDate = date
	newData.Details.Text = songDetails.Text
	newData.Details.Link = songDetails.Link
	debugParam := fmt.Sprintf("[DEBUG] newData.Group = %s | newData.Song = %s | newData.ReleaseDate = %s | newData.Text = %s | newData.Link = %s",
		newData.Main.Group, newData.Main.Song, newData.Details.ReleaseDate, newData.Details.Text, newData.Details.Link)
	logger.Log.Debug(debugParam)
	_, err = db.Db.Exec(`
		INSERT INTO music (song_id, release_date, text, link)
		VALUES (
			(SELECT song_id FROM songs WHERE song_name = $1 AND group_id = (SELECT group_id FROM groups WHERE group_name = $2)),
			TO_DATE($3,'DD-MM-YYYY'),
			$4,
			$5
		);
	`, newSong.Song, newSong.Group, newData.Details.ReleaseDate, newData.Details.Text, newData.Details.Link)

	if err != nil {
		logger.Log.Error("[ERROR] Ошибка добавления данных:", err)
		c.JSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"message": "Информация добавлена"})
	logger.Log.Info("[INFO] Запрос успешно выполнен")
}

// getSongDetailsFromAPI делает запрос в внешний АПИ для получения дополнительных данных (releaseDate, text, link)
func getSongDetailsFromAPI(group string, song string) *models.SongDetailRequest {
	logger.Log.Debug("[DEBUG] Вход в функцию getSongDetailsFromAPI")
	api_url := envutils.GetOutAPI()
	url := fmt.Sprintf(api_url, group, song)
	debugParam := fmt.Sprintf("[DEBUG] url = %s", url)
	logger.Log.Debug(debugParam)
	resp, err := http.Get(url)
	if err != nil {
		logger.Log.Error("[ERROR] Ошибка при получении данных песни: ", err)
		return nil
	}
	defer resp.Body.Close()
	debugParam = fmt.Sprintf("[DEBUG] resp.StatusCode = %d", resp.StatusCode)
	logger.Log.Debug(debugParam)
	if resp.StatusCode != 200 {
		logger.Log.Error("[ERROR] АПИ вернул не 200: ", resp.StatusCode)
		return nil
	}

	var songDetails models.SongDetailRequest
	if err := json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		logger.Log.Error("[ERROR] Ошибка при декодировании ответа АПИ: ", err)
		return nil
	}
	debugParam = fmt.Sprintf("[DEBUG] songDetails = %v", songDetails)
	logger.Log.Debug(debugParam)
	logger.Log.Info("[INFO] Запрос к АПИ успешно сделан")
	return &songDetails
}
