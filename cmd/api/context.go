package main

import (
	"context"
	"net/http"

	"github.com/TarekDTM/EnergyOptimizer/internal/data"
)

type ContextKey string

const userContextKey = ContextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)

	if !ok {
		panic("missing user value in request context")
	}
	return user
}
