package urls

import (
	"net/url"
)

// IsValidURL validates HTTP/HTTPS URLs only
func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
