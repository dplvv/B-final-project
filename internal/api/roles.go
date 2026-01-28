package api

import (
  "net/http"
)

func (api *API) RequireRole(roles ...string) func(http.Handler) http.Handler {
  allowed := map[string]bool{}
  for _, role := range roles {
    allowed[role] = true
  }

  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      userID := userIDFromContext(r.Context())
      if userID == "" {
        writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing user"})
        return
      }

      var role string
      err := api.DB.QueryRow("select role from users where id = $1", userID).Scan(&role)
      if err != nil {
        writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "user not found"})
        return
      }

      if !allowed[role] {
        writeJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
        return
      }

      next.ServeHTTP(w, r)
    })
  }
}

