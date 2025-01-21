package http

import (
	"crypto/rand"
	"math/big"
	"strings"
)

var (
	adjectives = []string{
		"quick", "silent", "autumn", "winter", "summer", "spring",
		"crystal", "azure", "coral", "crimson", "amber", "emerald",
		"gentle", "happy", "swift", "bright", "calm", "clever",
		"deep", "exotic", "fresh", "golden", "hidden", "infinite",
		"jovial", "lunar", "magical", "noble", "ocean", "peaceful",
		"radiant", "sacred", "scenic", "serene", "stellar", "cosmic",
	}

	nouns = []string{
		"river", "mountain", "forest", "meadow", "ocean", "desert",
		"valley", "garden", "island", "plateau", "canyon", "oasis",
		"harbor", "lagoon", "summit", "brook", "cascade", "dawn",
		"dusk", "echo", "flame", "flower", "glacier", "horizon",
		"lake", "meteor", "moon", "planet", "rain", "rainbow",
		"storm", "stream", "sunset", "thunder", "wave", "wind",
	}

	charset = strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")
)

func randChoice(arr []string) (string, error) {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(arr))))
	if err != nil {
		return "", err
	}
	return arr[i.Int64()], nil
}

func randString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		c, err := randChoice(charset)
		if err != nil {
			return "", err
		}
		b[i] = c[0]
	}
	return string(b), nil
}

func randPrefix() (string, error) {
	adj, err := randChoice(adjectives)
	if err != nil {
		return "", err
	}
	noun, err := randChoice(nouns)
	if err != nil {
		return "", err
	}
	suffix, err := randString(6)
	if err != nil {
		return "", err
	}
	return adj + "-" + noun + "-" + suffix, nil
}

// ParseDomain replaces a leading wildcard with a random string, or returns the domain as-is if no wildcard is present.
func ParseDomain(d string) (string, error) {
	if !strings.HasPrefix(d, "*.") {
		return d, nil
	}

	prefix, err := randPrefix()
	if err != nil {
		return "", err
	}
	return strings.Replace(d, "*", prefix, 1), nil
}
