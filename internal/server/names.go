package server

import (
	"crypto/rand"
	"fmt"
)

var adjectives = []string{"blue", "quiet", "small", "bright", "fast", "green"}
var nouns = []string{"river", "sun", "forest", "field", "stone", "cloud"}

func randomSubdomain() string {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "demo"
	}

	left := adjectives[int(b[0])%len(adjectives)]
	right := nouns[int(b[1])%len(nouns)]

	return fmt.Sprintf("%s-%s", left, right)
}
