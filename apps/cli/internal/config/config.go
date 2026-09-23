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

	DefaultURL = "https://folio.journ.app"
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

// Precedence: flag, environment, directory binding. Nothing global is
// persisted, so two shells can still hold different projects.
func Project(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv(EnvProject); env != "" {
		return env
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return bound(wd)
}

// Bind maps a directory and everything below it to a project. The map lives
// in the config dir rather than a dotfile so nothing lands in the repository.
func Bind(dir, slug string) error {
	key, err := canonical(dir)
	if err != nil {
		return err
	}
	bindings, _ := loadBindings()
	if bindings == nil {
		bindings = map[string]string{}
	}
	bindings[key] = slug
	return saveBindings(bindings)
}

func Unbind(dir string) error {
	key, err := canonical(dir)
	if err != nil {
		return err
	}
	bindings, err := loadBindings()
	if err != nil {
		return nil
	}
	delete(bindings, key)
	return saveBindings(bindings)
}

func bound(dir string) string {
	bindings, err := loadBindings()
	if err != nil || len(bindings) == 0 {
		return ""
	}
	current, err := canonical(dir)
	if err != nil {
		return ""
	}
	for {
		if slug, ok := bindings[current]; ok {
			return slug
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

// The same folder reached through a symlink (/var and /private/var on darwin)
// has to land on the same key, or a binding silently stops matching.
func canonical(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

func bindingsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects.json"), nil
}

func loadBindings() (map[string]string, error) {
	path, err := bindingsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func saveBindings(bindings map[string]string) error {
	path, err := bindingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(bindings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
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
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func configDir() (string, error) {
	// os.UserConfigDir ignores XDG_CONFIG_HOME on darwin, so tests need an
	// override that works on every platform.
	if dir := os.Getenv(EnvHome); dir != "" {
		return dir, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "folio"), nil
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
