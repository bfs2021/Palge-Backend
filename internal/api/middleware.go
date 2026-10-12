package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nongpal/Palge-Backend/internal/apperr"
	"github.com/nongpal/Palge-Backend/internal/audit"
	"github.com/nongpal/Palge-Backend/internal/token"
	"github.com/nongpal/Palge-Backend/internal/user"
	"github.com/nongpal/Palge-Backend/internal/validator"
)

type reqContextKey string

const requestIDKey reqContextKey = "requestID"

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, user.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.writeError(w, r, apperr.New(apperr.CodeInvalidToken))
			return
		}

		tkn := headerParts[1]

		v := validator.New()

		if token.ValidateTokenPlaintext(v, tkn); !v.Valid() {
			app.writeError(w, r, apperr.New(apperr.CodeInvalidToken))
			return
		}

		user, err := app.users.GetForToken(token.ScopeAuthentication, tkn)
		if err != nil {
			app.writeError(w, r, err)
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *Application) requiredAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.writeError(w, r, apperr.New(apperr.CodeUnauthenticated))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) requiredActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.writeError(w, r, apperr.New(apperr.CodeAccountInactive))
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requiredAuthenticatedUser(fn)
}

func (app *Application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		permission, err := app.permissions.GetAllForUser(user.ID)
		if err != nil {
			app.writeError(w, r, err)
			return
		}

		if !permission.Include(code) {
			app.recordAuditLog(r, &audit.AuditLog{
				UserID:     &user.ID,
				Action:     "authorize",
				Resource:   "account",
				ResourceID: nil,
				Result:     "failed",
			})
			app.writeError(w, r, apperr.New(apperr.CodePermissionDenied))
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.requiredActivatedUser(fn)
}

func (app *Application) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)

		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) contextGetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}

	return ""
}

func (app *Application) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(rw, r)

		app.logger.Info(
			"request completed",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"status", rw.status,
			"duration", time.Since(start),
			"request_id", app.contextGetRequestID(r.Context()),
			"ip", clientIP(r),
		)
	})
}
