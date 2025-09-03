package service

import (
	"errors"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"go-url-shortener/internal/storage"
)

// ShortenerService contains business logic for creating and resolving short URLs.
type ShortenerService struct {
	store storage.Store
	// domain represents the base URL used to format returned short links, e.g. https://sho.rt
	domain string
}

// New creates a ShortenerService with the provided storage and domain.
func New(store storage.Store, domain string) *ShortenerService {
	return &ShortenerService{store: store, domain: strings.TrimRight(domain, "/")}
}

// CreateShortRequest is the input for creating a short URL.
type CreateShortRequest struct {
	URL             string
	CustomAlias     string
	ExpireInSeconds int // optional: 0 or negative means no expiration
}

// CreateShortResponse is the output of creating a short URL.
type CreateShortResponse struct {
	Code  string `json:"code"`
	Short string `json:"short"`
}

// Stats describes analytics and metadata for a short code.
type Stats struct {
	Code        string     `json:"code"`
	URL         string     `json:"url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
	Clicks      int64      `json:"clicks"`
	LastAccess  *time.Time `json:"last_access,omitempty"`
	Expired     bool       `json:"expired"`
	ExpiresInMs *int64     `json:"expires_in_ms,omitempty"`
}

var (
	// ErrInvalidURL indicates the input is not an http(s) URL.
	ErrInvalidURL = errors.New("invalid url: must start with http or https")
	// ErrAliasTaken is returned when a requested alias is already in use.
	ErrAliasTaken = errors.New("custom alias already taken")
)

// CreateShort generates and stores a new short code for the given URL.
func (s *ShortenerService) CreateShort(req CreateShortRequest) (CreateShortResponse, error) {
	if !isValidHTTPURL(req.URL) {
		return CreateShortResponse{}, ErrInvalidURL
	}

	code := req.CustomAlias
	if code != "" {
		// Basic validation for custom alias characters
		if !isValidAlias(code) {
			return CreateShortResponse{}, errors.New("invalid custom alias: use alphanumerics, '-', '_' only")
		}
		if s.store.Exists(code) {
			return CreateShortResponse{}, ErrAliasTaken
		}
	} else {
		// generate a unique code
		var err error
		for i := 0; i < 5; i++ { // attempt a few times to avoid unlikely collision
			code, err = generateCode(6)
			if err != nil {
				return CreateShortResponse{}, err
			}
			if !s.store.Exists(code) {
				break
			}
		}
		if code == "" || s.store.Exists(code) {
			return CreateShortResponse{}, errors.New("could not allocate unique code")
		}
	}

	if err := s.store.Save(code, req.URL); err != nil {
		if errors.Is(err, storage.ErrCodeExists) {
			return CreateShortResponse{}, ErrAliasTaken
		}
		return CreateShortResponse{}, err
	}

	// optional expiry
	if req.ExpireInSeconds > 0 {
		exp := time.Now().Add(time.Duration(req.ExpireInSeconds) * time.Second)
		_ = s.store.SetExpire(code, exp)
	}

	short := s.domain + "/" + code
	return CreateShortResponse{Code: code, Short: short}, nil
}

// Resolve returns the original URL for a code if it exists and is not expired.
func (s *ShortenerService) Resolve(code string) (string, bool) {
	e, ok := s.store.GetEntry(code)
	if !ok {
		return "", false
	}
	if e.ExpireAt != nil && time.Now().After(*e.ExpireAt) {
		return "", false
	}
	return e.OriginalURL, true
}

// ResolveAndTrack returns the URL and records a click if found and active.
func (s *ShortenerService) ResolveAndTrack(code string) (string, bool) {
	url, ok := s.Resolve(code)
	if !ok {
		return "", false
	}
	_ = s.store.TrackHit(code)
	return url, true
}

// GetStats returns analytics and metadata for a code.
func (s *ShortenerService) GetStats(code string) (Stats, bool) {
	e, ok := s.store.GetEntry(code)
	if !ok {
		return Stats{}, false
	}
	st := Stats{
		Code:       code,
		URL:        e.OriginalURL,
		CreatedAt:  e.CreatedAt,
		ExpireAt:   e.ExpireAt,
		Clicks:     e.Clicks,
		LastAccess: e.LastAccess,
	}
	if e.ExpireAt != nil {
		expired := time.Now().After(*e.ExpireAt)
		st.Expired = expired
		if !expired {
			ms := e.ExpireAt.Sub(time.Now()).Milliseconds()
			st.ExpiresInMs = &ms
		}
	}
	return st, true
}

func isValidHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	return true
}

func isValidAlias(alias string) bool {
	for _, r := range alias {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return len(alias) > 0
}

// init seeds math/rand once for simple code generation.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// generateCode creates a random alphanumeric short code of the given length.
// Simpler for learning purposes; good enough for a demo/in-memory app.
func generateCode(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b), nil
}


