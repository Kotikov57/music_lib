//package models представляет собой пакет содержащий модели

package models

// Music структура, содержащая информацию о песне
type Music struct {
	Main    Song
	Details SongDetail
}

type Song struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
