// Package system provides the small driven adapters every wiring needs: a
// real clock, a PocketBase-compatible ID generator and an slog-backed logger.
package system

import (
	"crypto/rand"
	"log/slog"
	"time"
)

type Clock struct{}

func (Clock) Now() time.Time { return time.Now().UTC() }

// IDGenerator produces PocketBase record IDs: 15 lowercase alphanumerics,
// matching PocketBase's ^[a-z0-9]+$ default pattern. Generating the ID in the
// service rather than letting storage assign one means a record has a stable
// identity before it is written.
type IDGenerator struct{}

const idAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func (IDGenerator) NewID() string {
	b := make([]byte, 15)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is vanishingly unlikely, and a panic here would
		// take down a request for a cosmetic reason; fall back to the clock.
		nano := time.Now().UnixNano()
		for i := range b {
			b[i] = idAlphabet[int(nano>>(i%8))%len(idAlphabet)]
		}
		return string(b)
	}
	for i, v := range b {
		b[i] = idAlphabet[int(v)%len(idAlphabet)]
	}
	return string(b)
}

// TokenGenerator produces 43 alphanumerics, about 256 bits. Unlike
// IDGenerator it has no fallback: a token guessable from the clock would be
// worse than a failed request, and crypto/rand does not return errors.
type TokenGenerator struct{}

const tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (TokenGenerator) NewToken() string {
	out := make([]byte, 0, 43)
	buf := make([]byte, 64)
	for len(out) < 43 {
		rand.Read(buf)
		for _, v := range buf {
			// Bytes past the last whole multiple of 62 are dropped so every
			// character is equally likely.
			if v < 248 && len(out) < 43 {
				out = append(out, tokenAlphabet[v%62])
			}
		}
	}
	return string(out)
}

// Logger adapts log/slog to ports.Logger.
type Logger struct{ l *slog.Logger }

func NewLogger(l *slog.Logger) Logger {
	if l == nil {
		l = slog.Default()
	}
	return Logger{l: l}
}

func attrs(fields map[string]any) []any {
	out := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		out = append(out, k, v)
	}
	return out
}

func (g Logger) Debug(msg string, f map[string]any) { g.l.Debug(msg, attrs(f)...) }
func (g Logger) Info(msg string, f map[string]any)  { g.l.Info(msg, attrs(f)...) }
func (g Logger) Warn(msg string, f map[string]any)  { g.l.Warn(msg, attrs(f)...) }
func (g Logger) Error(msg string, f map[string]any) { g.l.Error(msg, attrs(f)...) }
