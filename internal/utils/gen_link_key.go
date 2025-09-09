package util

import (
	"math/rand"
)

func GenLinkKey() string {
	const (
		keyLength = 6
		// excluding certain characters to prevent generating inappropriate or hard-to-read strings:
		// vowels: a, e, i, o, u (including uppercase)
		// visually similar characters: O (uppercase), 0 (zero), l (lowercase L), and 1 (one), I (uppercase i)
		keyChars = "bcdfghjkmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ23456789"
	)

	result := make([]byte, 0, keyLength)
	for i := 0; i < keyLength; i++ {
		randIdx := rand.Intn(len(keyChars))
		result = append(result, keyChars[randIdx])
	}

	return string(result)
}
