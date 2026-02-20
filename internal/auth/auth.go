package auth

import (
	"errors"
	"strings"
)

type Claims struct {
	TenantID string
	Subject  string
}

type Authenticator struct {
	tokens map[string]Claims
}

func New(raw string) *Authenticator {
	tokens := map[string]Claims{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, ":")
		if len(fields) < 2 {
			continue
		}
		token := strings.TrimSpace(fields[0])
		tenantID := strings.TrimSpace(fields[1])
		subject := "unknown"
		if len(fields) >= 3 {
			subject = strings.TrimSpace(fields[2])
		}
		if token == "" || tenantID == "" {
			continue
		}
		tokens[token] = Claims{TenantID: tenantID, Subject: subject}
	}
	return &Authenticator{tokens: tokens}
}

var (
	ErrMissingAuthHeader = errors.New("missing_authorization_header")
	ErrInvalidAuthScheme = errors.New("invalid_authorization_scheme")
	ErrInvalidToken      = errors.New("invalid_token")
)

func (a *Authenticator) Authenticate(authHeader string) (Claims, error) {
	if strings.TrimSpace(authHeader) == "" {
		return Claims{}, ErrMissingAuthHeader
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return Claims{}, ErrInvalidAuthScheme
	}
	token := strings.TrimSpace(parts[1])
	claims, ok := a.tokens[token]
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}
