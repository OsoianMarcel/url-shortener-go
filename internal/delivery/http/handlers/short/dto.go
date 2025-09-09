package short

import "time"

type shortenRequestDTO struct {
	URL string `json:"url"`
}

type shortenResponseDTO struct {
	ShortURL string `json:"short_url"`
	Key      string `json:"key"`
}

type expandResponseDTO struct {
	URL string `json:"url"`
}

type statsResponseDTO struct {
	Hits      uint      `json:"hits"`
	CreatedAt time.Time `json:"created_at"`
}
