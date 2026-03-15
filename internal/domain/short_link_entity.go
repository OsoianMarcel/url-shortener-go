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

func NewShortLink(key, originalURL string) ShortLink {
	return ShortLink{
		Key:         key,
		OriginalURL: originalURL,
		Hits:        0,
		CreatedAt:   time.Now(),
	}
}
