package api

import (
	"context"
	"net/http"

	"github.com/nongpal/Palge-Backend/internal/user"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *Application) contextSetUser(r *http.Request, user *user.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *Application) contextGetUser(r *http.Request) *user.User {
	user, ok := r.Context().Value(userContextKey).(*user.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
