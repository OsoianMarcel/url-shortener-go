package short

import "errors"

var (
	ErrShortLinkNotFound = errors.New("short link not found")
	ErrInvalidURL        = errors.New("invalid URL")
)
