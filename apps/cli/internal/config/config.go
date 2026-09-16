package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvURL     = "FOLIO_URL"
	EnvToken   = "FOLIO_TOKEN"
	EnvProject = "FOLIO_PROJECT"
	EnvHome    = "FOLIO_CONFIG_DIR"

	DefaultURL = "http://127.0.0.1:8090"
)

var ErrNoToken = errors.New("no token: set " + EnvToken + " or run `folio login`")

type Config struct {
	URL   string
	Token string
}

type cached struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Not persisted: two shells can hold different projects.
func Project(flag string) string {
	if flag != "" {
		return flag
	}
	return os.Getenv(EnvProject)
}

func SetURL(raw string) error {
	normalized, err := NormalizeURL(raw)
	if err != nil {
		return err
	}

	stored, _ := load()
	return Save(normalized, stored.Token)
}

func Path() (string, error) {
	// os.UserConfigDir ignores XDG_CONFIG_HOME on darwin, so tests need an
	// override that works on every platform.
	if dir := os.Getenv(EnvHome); dir != "" {
		return filepath.Join(dir, "credentials.json"), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "folio", "credentials.json"), nil
}

// NormalizeURL turns what someone actually types into an origin the client can
// append paths to. A bare host is the common case when pasting a deployment
// address, and a pasted URL often carries a path that would otherwise produce
// a 404 far from its cause.
func NormalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("empty URL")
	}

	if !strings.Contains(trimmed, "://") {
		// Loopback almost never has TLS; anything else is assumed to.
		scheme := "https"
		if host, _, _ := strings.Cut(trimmed, ":"); host == "localhost" || host == "127.0.0.1" || host == "[::1]" {
			scheme = "http"
		}
		trimmed = scheme + "://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid URL %q: scheme must be http or https, got %q", raw, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid URL %q: missing host", raw)
	}

	return parsed.Scheme + "://" + parsed.Host, nil
}

// Precedence: flag, environment, cached login.
func Resolve(urlFlag, tokenFlag string) (Config, error) {
	cfg := Config{URL: urlFlag, Token: tokenFlag}

	if cfg.URL == "" {
		cfg.URL = os.Getenv(EnvURL)
	}
	if cfg.Token == "" {
		cfg.Token = os.Getenv(EnvToken)
	}

	// The cache is a pair: a token is only valid for the host that issued it.
	// Taking one half alongside a token from elsewhere sends a credential to a
	// server that never issued it, which surfaces as a 401 far from its cause.
	if cfg.URL == "" && cfg.Token == "" {
		if stored, err := load(); err == nil {
			cfg.URL, cfg.Token = stored.URL, stored.Token
		}
	}

	if cfg.URL == "" {
		cfg.URL = DefaultURL
	}

	normalized, err := NormalizeURL(cfg.URL)
	if err != nil {
		return Config{}, err
	}
	cfg.URL = normalized

	return cfg, nil
}

func load() (cached, error) {
	path, err := Path()
	if err != nil {
		return cached{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cached{}, err
	}
	var out cached
	if err := json.Unmarshal(data, &out); err != nil {
		return cached{}, err
	}
	return out, nil
}

func Save(url, token string) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(cached{URL: url, Token: token})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
