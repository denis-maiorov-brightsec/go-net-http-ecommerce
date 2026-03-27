package auth

import (
	"net/http"
	"strings"
)

const ManagePromotionsPermission = "promotions:manage"

type StubAuthenticator struct {
	Tokens map[string]Principal
}

func DefaultStubAuthenticator() StubAuthenticator {
	return StubAuthenticator{
		Tokens: map[string]Principal{
			"promotions-admin": NewPrincipal("promotions-admin", ManagePromotionsPermission),
			"catalog-readonly": NewPrincipal("catalog-readonly"),
		},
	}
}

func (a StubAuthenticator) Authenticate(r *http.Request) (Principal, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return Principal{}, ErrUnauthorized
	}

	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(strings.TrimSpace(scheme), "Bearer") {
		return Principal{}, ErrUnauthorized
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return Principal{}, ErrUnauthorized
	}

	principal, found := a.Tokens[token]
	if !found {
		return Principal{}, ErrUnauthorized
	}

	return principal, nil
}
