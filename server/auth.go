package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"regexp"
	"strings"

	"github.com/equals215/deepsentinel/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

var (
	apiProtectedURLs = []*regexp.Regexp{
		regexp.MustCompile("^/probe(/.*)?$"),
	}
	dashboardProtectedURLs = []*regexp.Regexp{
		regexp.MustCompile("^/dashboard$"),
	}
	dashboardWSprotectedURLs = []*regexp.Regexp{
		regexp.MustCompile("^/dashws(/.*)?$"),
	}
)

func authFilterAPI(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	for _, pattern := range apiProtectedURLs {
		if pattern.MatchString(originalURL) {
			return false
		}
	}
	return true
}

func authFilterDashboardWS(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	for _, pattern := range dashboardWSprotectedURLs {
		if pattern.MatchString(originalURL) {
			return false
		}
	}
	return true
}

func authFilterDashboard(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	for _, pattern := range dashboardProtectedURLs {
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

func fiberSetAuth(app *fiber.App) {
	app.Use(keyauth.New(keyauth.Config{
		Next:      authFilterAPI,
		KeyLookup: "header:Authorization",
		Validator: validateAuth,
	}))

	app.Use(basicauth.New(basicauth.Config{
		Next:  authFilterDashboardWS,
		Realm: "Dashboard",
		Authorizer: func(user, pass string) bool {
			if user == "admin" {
				ok, err := validateAuth(nil, pass)
				if err != nil {
					return false
				}
				return ok
			}
			return false
		},
	}))

	app.Use(keyauth.New(keyauth.Config{
		Next:      authFilterDashboard,
		KeyLookup: "cookie:auth_token",
		Validator: validateAuth,
	}))
}
