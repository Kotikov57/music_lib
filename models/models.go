//package models представляет собой пакет содержащий модели

package models

import(
	"time"
)

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
	ReleaseDate time.Time `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

//MusicRequest специальная структура введённая для преобразования даты к формату "dd-mm-yyyy"
type MusicRequest struct {
	Main    Song
	Details SongDetailRequest
}

//SongDetailRequest специальная структура введённая для преобразования даты к формату "dd-mm-yyyy"
type SongDetailRequest struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
