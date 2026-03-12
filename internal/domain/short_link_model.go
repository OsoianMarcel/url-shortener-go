package domain

import "time"

type CreateAction struct {
	OriginalURL string
}

type CreateResult struct {
	ShortURL string
	Key      string
}

type StatsResult struct {
	Hits      uint
	CreatedAt time.Time
}
