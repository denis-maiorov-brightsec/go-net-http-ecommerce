package auth

import (
	"errors"
	"net/http"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type Principal struct {
	Subject     string
	Permissions map[string]struct{}
}

func NewPrincipal(subject string, permissions ...string) Principal {
	grants := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		grants[permission] = struct{}{}
	}

	return Principal{
		Subject:     subject,
		Permissions: grants,
	}
}

func (p Principal) HasPermission(permission string) bool {
	if permission == "" {
		return true
	}

	_, ok := p.Permissions[permission]
	return ok
}

type Authenticator interface {
	Authenticate(*http.Request) (Principal, error)
}

type Middleware struct {
	authenticator Authenticator
}

func NewMiddleware(authenticator Authenticator) *Middleware {
	return &Middleware{authenticator: authenticator}
}

func (m *Middleware) Require(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m == nil || m.authenticator == nil {
				apierror.Write(w, r, apierror.Internal(errors.New("authenticator is not configured")))
				return
			}

			principal, err := m.authenticator.Authenticate(r)
			if err != nil {
				switch {
				case errors.Is(err, ErrForbidden):
					apierror.Write(w, r, apierror.Forbidden("Forbidden"))
				case errors.Is(err, ErrUnauthorized):
					apierror.Write(w, r, apierror.Unauthorized("Authentication required"))
				default:
					apierror.Write(w, r, apierror.Internal(err))
				}
				return
			}

			if !principal.HasPermission(permission) {
				apierror.Write(w, r, apierror.Forbidden("Forbidden"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
