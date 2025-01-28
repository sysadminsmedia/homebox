package v1

import (
	"net/url"

	"github.com/rs/zerolog/log"
)

func GetHBURL(refererHeader, fallback string) (hbURL string) {
	hbURL = refererHeader
	if hbURL == "" {
		hbURL = fallback
	}

	return stripPathFromURL(hbURL)
}

// stripPathFromURL removes the path from a URL.
// ex. https://example.com/tools -> https://example.com
func stripPathFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Err(err).Msg("failed to parse URL")
		return ""
	}

	strippedURL := url.URL{Scheme: parsedURL.Scheme, Host: parsedURL.Host}

	return strippedURL.String()
}
