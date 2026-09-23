package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  string
		error bool
	}{
		{name: "leaves an https origin alone", in: "https://folio.example.com", want: "https://folio.example.com"},
		{name: "strips a trailing slash", in: "https://folio.example.com/", want: "https://folio.example.com"},
		{name: "keeps an explicit port", in: "http://127.0.0.1:8090", want: "http://127.0.0.1:8090"},
		{
			name: "assumes https for a bare remote host",
			in:   "folio.example.com",
			want: "https://folio.example.com",
		},
		{
			name: "assumes http for a bare loopback host, which rarely has TLS",
			in:   "127.0.0.1:8090",
			want: "http://127.0.0.1:8090",
		},
		{name: "assumes http for bare localhost", in: "localhost:8090", want: "http://localhost:8090"},
		{
			name: "drops a path, since the client appends its own",
			in:   "https://folio.example.com/api",
			want: "https://folio.example.com",
		},
		{name: "trims surrounding whitespace", in: "  https://folio.example.com  ", want: "https://folio.example.com"},
		{name: "rejects an unsupported scheme", in: "htp://folio.example.com", error: true},
		{name: "rejects a non-http scheme", in: "ftp://folio.example.com", error: true},
		{name: "rejects an empty host", in: "https://", error: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NormalizeURL(c.in)

			if c.error {
				if err == nil {
					t.Fatalf("NormalizeURL(%q) error = nil, want an error", c.in)
				}
				return
			}

			if err != nil {
				t.Fatalf("NormalizeURL(%q) error = %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("NormalizeURL(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestResolvePrecedence(t *testing.T) {
	t.Setenv(EnvURL, "https://from-env.example.com")
	t.Setenv(EnvToken, "env-token")

	t.Run("the flag beats the environment", func(t *testing.T) {
		cfg, err := Resolve("https://from-flag.example.com", "")
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if cfg.URL != "https://from-flag.example.com" {
			t.Errorf("URL = %q", cfg.URL)
		}
		if cfg.Token != "env-token" {
			t.Errorf("Token = %q, want the environment token", cfg.Token)
		}
	})

	t.Run("the environment is normalized too", func(t *testing.T) {
		t.Setenv(EnvURL, "from-env.example.com/")
		cfg, err := Resolve("", "")
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if cfg.URL != "https://from-env.example.com" {
			t.Errorf("URL = %q, want the normalized environment URL", cfg.URL)
		}
	})

	t.Run("a malformed URL is reported rather than sent", func(t *testing.T) {
		if _, err := Resolve("htp://nope.example.com", ""); err == nil {
			t.Fatal("Resolve() error = nil, want an error for a bad scheme")
		}
	})
}

func TestResolveNeverMixesACachedURLWithAnotherToken(t *testing.T) {
	t.Setenv(EnvURL, "")
	t.Setenv(EnvToken, "token-from-env")
	t.Setenv(EnvHome, t.TempDir())

	if err := Save("https://remote.example.com", "token-from-remote"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if cfg.Token == "token-from-env" && cfg.URL == "https://remote.example.com" {
		t.Error("paired a cached remote URL with the environment token")
	}
	if cfg.URL != DefaultURL {
		t.Errorf("URL = %q, want the default once a token comes from elsewhere", cfg.URL)
	}
}

func TestResolveUsesBothHalvesOfTheCacheTogether(t *testing.T) {
	t.Setenv(EnvURL, "")
	t.Setenv(EnvToken, "")
	t.Setenv(EnvHome, t.TempDir())

	if err := Save("https://remote.example.com", "token-from-remote"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if cfg.URL != "https://remote.example.com" || cfg.Token != "token-from-remote" {
		t.Errorf("got %q / %q, want the cached pair", cfg.URL, cfg.Token)
	}
}

func TestProject(t *testing.T) {
	t.Run("reads the environment", func(t *testing.T) {
		t.Setenv(EnvProject, "folio")
		if got := Project(""); got != "folio" {
			t.Errorf("Project(\"\") = %q, want folio", got)
		}
	})

	t.Run("a flag overrides the environment", func(t *testing.T) {
		t.Setenv(EnvProject, "folio")
		if got := Project("other"); got != "other" {
			t.Errorf("Project(\"other\") = %q, want other", got)
		}
	})

	t.Run("is empty when nothing is set, so callers can require it", func(t *testing.T) {
		t.Setenv(EnvProject, "")
		t.Setenv(EnvHome, t.TempDir())
		t.Chdir(t.TempDir())
		if got := Project(""); got != "" {
			t.Errorf("Project(\"\") = %q, want empty", got)
		}
	})
}

func TestSetURLPersistsWithoutTouchingTheToken(t *testing.T) {
	t.Setenv(EnvURL, "")
	t.Setenv(EnvToken, "")
	t.Setenv(EnvHome, t.TempDir())

	if err := Save("https://old.example.com", "keep-me"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := SetURL("new.example.com"); err != nil {
		t.Fatalf("SetURL() error = %v", err)
	}

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if cfg.URL != "https://new.example.com" {
		t.Errorf("URL = %q, want the normalized new URL", cfg.URL)
	}
	if cfg.Token != "keep-me" {
		t.Errorf("Token = %q, want the existing token preserved", cfg.Token)
	}
}

func TestSetURLRejectsAMalformedURL(t *testing.T) {
	t.Setenv(EnvHome, t.TempDir())

	if err := SetURL("htp://nope.example.com"); err == nil {
		t.Fatal("SetURL() error = nil, want an error for a bad scheme")
	}
}

func TestResolveFallsBackToTheDefault(t *testing.T) {
	t.Setenv(EnvURL, "")
	t.Setenv(EnvToken, "tok")

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if cfg.URL != DefaultURL {
		t.Errorf("URL = %q, want %q", cfg.URL, DefaultURL)
	}
}

func TestProjectFromABoundDirectory(t *testing.T) {
	t.Setenv(EnvProject, "")
	t.Setenv(EnvHome, t.TempDir())

	root := t.TempDir()
	nested := filepath.Join(root, "apps", "cli")
	inner := filepath.Join(root, "vendor")
	for _, dir := range []string{nested, inner} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := Bind(root, "folio"); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if err := Bind(inner, "other"); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	t.Run("the bound directory resolves", func(t *testing.T) {
		t.Chdir(root)
		if got := Project(""); got != "folio" {
			t.Errorf("Project(\"\") = %q, want folio", got)
		}
	})

	t.Run("a subdirectory inherits the binding", func(t *testing.T) {
		t.Chdir(nested)
		if got := Project(""); got != "folio" {
			t.Errorf("Project(\"\") = %q, want folio", got)
		}
	})

	t.Run("the nearest binding wins", func(t *testing.T) {
		t.Chdir(inner)
		if got := Project(""); got != "other" {
			t.Errorf("Project(\"\") = %q, want other", got)
		}
	})

	t.Run("the environment beats the binding", func(t *testing.T) {
		t.Chdir(root)
		t.Setenv(EnvProject, "from-env")
		if got := Project(""); got != "from-env" {
			t.Errorf("Project(\"\") = %q, want from-env", got)
		}
	})

	t.Run("an unbound directory resolves nothing", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if got := Project(""); got != "" {
			t.Errorf("Project(\"\") = %q, want empty", got)
		}
	})

	t.Run("unbinding leaves the other bindings", func(t *testing.T) {
		if err := Unbind(inner); err != nil {
			t.Fatalf("Unbind() error = %v", err)
		}
		t.Chdir(inner)
		if got := Project(""); got != "folio" {
			t.Errorf("Project(\"\") = %q, want the parent's folio", got)
		}
	})
}

func TestBindDoesNotTouchCredentials(t *testing.T) {
	t.Setenv(EnvURL, "")
	t.Setenv(EnvToken, "")
	t.Setenv(EnvHome, t.TempDir())

	if err := Save("https://remote.example.com", "keep-me"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := Bind(t.TempDir(), "folio"); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if cfg.Token != "keep-me" {
		t.Errorf("Token = %q, want the cached token untouched", cfg.Token)
	}
}
