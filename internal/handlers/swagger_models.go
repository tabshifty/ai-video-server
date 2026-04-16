package handlers

type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type AuthRegisterRequest struct {
	Username string `json:"username" example:"alice"`
	Email    string `json:"email" example:"alice@example.com"`
	Password string `json:"password" example:"secret123"`
}

type AuthLoginRequest struct {
	Username string `json:"username" example:"alice"`
	Email    string `json:"email" example:"alice@example.com"`
	Password string `json:"password" example:"secret123"`
}

type UploadCheckRequest struct {
	Hash     string `json:"hash" example:"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"`
	FileSize int64  `json:"file_size" example:"10485760"`
}

type AdminScrapePreviewRequest struct {
	VideoID string `json:"video_id" example:"8fbe33de-0189-4972-a7c5-b77c9e75b859"`
	Title   string `json:"title" example:"Inception"`
	Year    int    `json:"year" example:"2010"`
	Type    string `json:"type" example:"movie"`
}

type AdminScrapeConfirmRequest struct {
	VideoID       string                 `json:"video_id" example:"8fbe33de-0189-4972-a7c5-b77c9e75b859"`
	TMDBID        int                    `json:"tmdb_id" example:"27205"`
	Title         string                 `json:"title" example:"盗梦空间"`
	Overview      string                 `json:"overview" example:"自定义简介"`
	PosterURL     string                 `json:"poster_url" example:"https://image.tmdb.org/t/p/original/xxx.jpg"`
	ReleaseDate   string                 `json:"release_date" example:"2010-07-16"`
	Metadata      map[string]interface{} `json:"metadata"`
	SeasonNumber  int                    `json:"season_number" example:"1"`
	EpisodeNumber int                    `json:"episode_number" example:"2"`
}

type AdminActorScrapePreviewRequest struct {
	Name   string `json:"name" example:"三上悠亚"`
	Source string `json:"source" example:"javdb"`
	Limit  int    `json:"limit" example:"10"`
}
