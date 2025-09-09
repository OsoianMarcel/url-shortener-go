package entities

import "time"

type ShortLink struct {
	// Primary key
	ID string
	// The unique key
	Key         string
	OriginalURL string
	ShortURL    string
	Hits        uint
	CreatedAt   time.Time
}
