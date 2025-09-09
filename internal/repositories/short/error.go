package short

import "errors"

var (
	ErrShortLinkNotFound  = errors.New("short link not found")
	ErrShortLinkKeyExists = errors.New("short link key exists")
)
