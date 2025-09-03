package urls

import (
	"net/url"
)

// IsValidURL validates that the URL is a valid HTTP or HTTPS URL
func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
