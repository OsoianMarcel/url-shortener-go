package domain

import "time"

type ShortLink struct {
	// Primary key
	ID string
	// The unique key
	Key         string
	OriginalURL string
	Hits        uint
	CreatedAt   time.Time
}
