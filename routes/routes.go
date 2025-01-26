//package routes представляет собой пакет для работы с маршрутами

package routes

import (
	"database/sql"
	"effect_mobile/db"
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
	log.Println("[DEBUG] Вход в функцию GetData")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	groupName := c.Query("group")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	keyword := c.Query("keyword")
	log.Printf("[DEBUG] page = %d | limit = %d | offset = %d", page, limit, offset)
	query := "SELECT groups.group_name AS group, songs.song_name AS song, release_date, text, link" + 
	"FROM details" +
	"JOIN groups ON details.group_id = groups.group_id" +
	"JOIN songs ON details.song_id = songs.song_id" +
	"WHERE ($1 = '' OR groups.group_name = $1) " +
	"AND ($2 = '' OR details.release_date >= $2) AND ($3 = '' OR m.release_date <= $3) " +
	"AND ($4 = '' OR details.text ILIKE '%' || $4 || '%')" + 
	"ORDER BY details.release_date DESC;" +
	"LIMIT $5 OFFSET $6"
	
	rows, err := db.Db.Query(query, groupName, startDate, endDate, keyword, limit, offset)

	if err != nil {
		log.Println("[ERROR] Ошибка выполнения запроса:", err)
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
			log.Println("[ERROR] Ошибка сканирования строки:", err)
			c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
			return
		}
		req.Main.Group = m.Main.Group
		req.Main.Song = m.Main.Song
		req.Details.ReleaseDate = m.Details.ReleaseDate.Format("02-01-2006")
		req.Details.Text = m.Details.Text
		req.Details.Link = m.Details.Link
		log.Printf("[DEBUG] Group = %s | Song = %s | ReleaseDate = %s | Text = %s | Link = %s",
		req.Main.Group, req.Main.Song, req.Details.ReleaseDate, req.Details.Text, req.Details.Link)
		data = append(data, req)
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
	songIdParam := c.Query("song")
	log.Printf("[DEBUG] songParam = %s", songIdParam)

	var text string
	err := db.Db.QueryRow(`SELECT text FROM details` +
	`WHERE song_id = $1`, songIdParam).Scan(&text)
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
	songIdParam := c.Query("song")
	log.Printf("[DEBUG] songParam = %s", songIdParam)

	tx, err := db.Db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error" : "Не получилось начать транзакцию"})
	}
	defer tx.Rollback()
	result1, err := tx.Exec(`DELETE FROM details WHERE song_id = $1`, songIdParam)
	if err != nil {
		log.Println("[ERROR] Ошибка при удалении записи из details:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}
	result2, err := tx.Exec(`DELETE FROM songs WHERE song_id = $1`, songIdParam)
	if err != nil {
		log.Println("[ERROR] Ошибка при удалении записи из songs:", err)
		c.JSON(500, gin.H{"error": "Ошибка базы данных"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(500, gin.H{"error": "Не удалось закоммитить транзакцию"})
		return
	}
	rowsAffected1, err := result1.RowsAffected()
	if err != nil {
		log.Println("[ERROR] Ошибка получения затронутых строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
		return
	}
	rowsAffected2, err := result2.RowsAffected()
	if err != nil {
		log.Println("[ERROR] Ошибка получения затронутых строк:", err)
		c.JSON(500, gin.H{"error": "Ошибка обработки данных"})
		return
	}
	if rowsAffected1 == 0 || rowsAffected2 == 0 {
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
	songIdParam := c.Query("song")
	log.Printf("[DEBUG] songParam = %s", songIdParam)

	var updatedDataRequest models.MusicRequest
	if err := c.ShouldBindJSON(&updatedDataRequest); err != nil {
		log.Println("[ERROR] Неверный формат данных:", err)
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}
	tx, err := db.Db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error" : "Не получилось начать транзакцию"})
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE songs SET song_name = $1 WHERE song_id = $2", updatedDataRequest.Main.Song, songIdParam)
	if err != nil {
		c.JSON(500, gin.H{"error": "Не удалось обновить название песни"})
		return
	}
	date, err := time.Parse("02-01-2006", updatedDataRequest.Details.ReleaseDate)
	if err != nil {
		c.JSON(500, gin.H{"error": "Не удалось преобразовать формат времени"})
	}
	_, err = tx.Exec("UPDATE details SET release_date = $1, text = $2, link = $3 WHERE song_id = $4",
	 date, updatedDataRequest.Details.Text, updatedDataRequest.Details.Link, songIdParam)
	
	 if err != nil {
		c.JSON(500, gin.H{"error": "Не удалось обновить данные песни"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(500, gin.H{"error": "Не удалось закоммитить транзакцию"})
		return
	}

	c.JSON(200, gin.H{"message": "Информация обновлена"})
	log.Println("[INFO] Запрос успешно выполнен")
}

func PutParam( c *gin.Context) {
	log.Println("[DEBUG] Вход в функцию PutData")
	songIdParam := c.Query("song")
	log.Printf("[DEBUG] songParam = %s", songIdParam)
	param := c.Query("param")
	
	switch param {
	case "name":
		var name string
		if err := c.ShouldBindJSON(&name); err != nil {
			log.Println("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE songs SET song_name = $1 WHERE song_id = $2", name, songIdParam)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось обновить название песни"})
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
	case "release_date":
		var date string
		if err := c.ShouldBindJSON(&date); err != nil {
			log.Println("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}

		dateFormatted, err := time.Parse("02-01-2006", date)
		if err != nil {
			log.Println("[ERROR] Неверный формат времени:", err)
			c.JSON(400, gin.H{"error": "Неверный формат времени"})
			return
		}

		result, err := db.Db.Exec("UPDATE details SET release_id = $1 WHERE song_id = $2", dateFormatted, songIdParam)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось обновить дату релиза песни"})
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
	case "text":
		var text string
		if err := c.ShouldBindJSON(&text); err != nil {
			log.Println("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE details SET text = $1 WHERE song_id = $2", text, songIdParam)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось обновить текст песни"})
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
	case "link":
		var link string
		if err := c.ShouldBindJSON(&link); err != nil {
			log.Println("[ERROR] Неверный формат данных:", err)
			c.JSON(400, gin.H{"error": "Неверный формат данных"})
			return
		}
		result, err := db.Db.Exec("UPDATE details SET link = $1 WHERE song_id = $2", link, songIdParam)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось обновить ссылку на песню"})
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
	default:
		log.Println("[ERROR] Некорректный параметр")
		c.JSON(400, gin.H{"error": "Некорректный параметр"})
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
	log.Println("[DEBUG] Вход в функцию PostData")
	var newSong models.Song
	if err := c.ShouldBindJSON(&newSong); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	_, err := db.Db.Exec(`
		INSERT INTO groups (group_name)
		VALUES ($1)
		ON CONFLICT (group_name) DO NOTHING;
	`, newSong.Group)
	if err != nil {
		log.Println("[ERROR] Ошибка при добавлении группы")
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
		log.Println("[ERROR] Ошибка при добавлении песни")
		c.JSON(500, gin.H{"error": "Ошибка при добавлении песни"})
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
	date, err := time.Parse("02-01-2006", songDetails.ReleaseDate)
	if err != nil {
		log.Println("[ERROR] Неверный формат времени:", err)
		c.JSON(400, gin.H{"error": "Неверный формат времени"})
		return
	}
	newData.Details.ReleaseDate = date
	newData.Details.Text = songDetails.Text
	newData.Details.Link = songDetails.Link
	log.Printf("[DEBUG] newData.Group = %s | newData.Song = %s | newData.ReleaseDate = %s | newData.Text = %s | newData.Link = %s", newData.Main.Group, newData.Main.Song, newData.Details.ReleaseDate, newData.Details.Text, newData.Details.Link)
	_, err = db.Db.Exec(`
		INSERT INTO music (song_id, release_date, text, link)
		VALUES (
			(SELECT song_id FROM songs WHERE song_name = $1 AND group_id = (SELECT group_id FROM groups WHERE group_name = $2)),
			$3,
			$4,
			$5
		);
	`, newSong.Song, newSong.Group, newData.Details.ReleaseDate, newData.Details.Text, newData.Details.Link)

	if err != nil {
		log.Println("[ERROR] Ошибка добавления данных:", err)
		c.JSON(500, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"message": "Информация добавлена"})
	log.Println("[INFO] Запрос успешно выполнен")
}

// getSongDetailsFromAPI делает запрос в внешний АПИ для получения дополнительных данных (releaseDate, text, link)
func getSongDetailsFromAPI(group string, song string) *models.SongDetailRequest {
	log.Println("[DEBUG] Вход в функцию getSongDetailsFromAPI")
	api_url := envutils.GetOutAPI()
	url := fmt.Sprintf(api_url, group, song)
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

	var songDetails models.SongDetailRequest
	if err := json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		log.Println("[ERROR] Ошибка при декодировании ответа АПИ: ", err)
		return nil
	}
	log.Printf("[DEBUG] songDetails = %v", songDetails)
	log.Println("[INFO] Запрос к АПИ успешно сделан")
	return &songDetails
}
