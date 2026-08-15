package httpx

import (
	"context"
	"net"
	"net/http"

	"curanmor-ai/internal/auth"
)

type ctxKey int

const (
	ctxKeyClaims ctxKey = iota
	ctxKeyAPIKeyID
)

func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, ctxKeyClaims, c)
}

func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	c, ok := ctx.Value(ctxKeyClaims).(*auth.Claims)
	return c, ok
}

func WithAPIKeyID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, ctxKeyAPIKeyID, id)
}

func APIKeyIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxKeyAPIKeyID).(int64)
	return id, ok
}

// ClientIP mengekstrak IP klien, mempertimbangkan header X-Forwarded-For
// yang diset oleh reverse proxy (nginx) di depan API.
func ClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
