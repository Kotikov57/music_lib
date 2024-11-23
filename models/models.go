//package models представляет собой пакет содержащий модели

package models

// Music структура, содержащая информацию о песне
type Music struct {
	Group       string `json:"group"`
	Song        string `json:"song"`
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
