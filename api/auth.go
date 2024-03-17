package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"regexp"
	"strings"

	"github.com/equals215/deepsentinel/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

var protectedURLs = []*regexp.Regexp{
	regexp.MustCompile("^/report$"),
}

func authFilter(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	for _, pattern := range protectedURLs {
		if pattern.MatchString(originalURL) {
			return false
		}
	}
	return true
}

func validateAuth(_ *fiber.Ctx, givenKey string) (bool, error) {
	hashedKey := sha256.Sum256([]byte(config.Server.AuthToken))
	hashedGivenKey := sha256.Sum256([]byte(givenKey))

	if subtle.ConstantTimeCompare(hashedKey[:], hashedGivenKey[:]) == 1 {
		return true, nil
	}
	return false, keyauth.ErrMissingOrMalformedAPIKey
}
