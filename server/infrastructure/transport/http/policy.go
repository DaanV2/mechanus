package http

import (
	"errors"
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/authentication"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

// RequirePolicy guards a route with a policy. A credential that isn't accepted gets a 401, even on an anonymous route.
// So does no identity on a route that needs one, an identity that isn't allowed gets a 403.
// The identity is put on the context of the request, see [authentication.IdentityFromContext].
func RequirePolicy(identities *authentication.IdentityResolver, policy authentication.Policy, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := identities.Resolve(r.Context(), r.Header)
		if err != nil {
			logging.From(r.Context()).Info("rejected credential", "policy", policy, "error", err)
			http.Error(w, "Could not authenticate: "+err.Error(), http.StatusUnauthorized)

			return
		}

		if err := policy.Authorize(identity); err != nil {
			status := http.StatusForbidden
			if errors.Is(err, authentication.ErrIdentityRequired) {
				status = http.StatusUnauthorized
			}
			logging.From(r.Context()).Info("route policy denied", "policy", policy, "identity", identity.Kind, "error", err)
			http.Error(w, err.Error(), status)

			return
		}

		next.ServeHTTP(w, r.WithContext(authentication.ContextWithIdentity(r.Context(), identity)))
	})
}
