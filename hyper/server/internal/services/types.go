package services

type TorrentSearchResult struct {
	InfoHash string
	Name     string
}

type TorrentioSource struct {
	Name          string `json:"name"`
	Title         string `json:"title"`
	InfoHash      string `json:"infoHash"`
	FileIdx       int    `json:"fileIdx"`
	BehaviorHints struct {
		BingeGroup string `json:"bingeGroup"`
		Filename   string `json:"filename"`
	} `json:"behaviorHints"`
	Sources []string `json:"sources"`
}

type TorrentionResponse struct {
	Streams []TorrentioSource `json:"streams"`
}

type DiscoverMoviesResp struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	PosterPath  string  `json:"poster_path"`
	Overview    string  `json:"overview"`
	Language    string  `json:"original_language,omitempty"`
	VoteAverage float64 `json:"vote_average"`
	GenreIDs    []int   `json:"genre_ids,omitempty"`
	IsWatched   bool    `json:"isWatched"`
}
