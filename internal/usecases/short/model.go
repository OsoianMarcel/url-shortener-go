package short

import "time"

type CreateInput struct {
	OriginalURL string
}

type CreateOutput struct {
	ShortURL string
	Key      string
}

type StatsOutput struct {
	Hits      uint
	CreatedAt time.Time
}
