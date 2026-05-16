package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/TarekDTM/EnergyOptimizer/internal/data"
	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var input struct {
		Email             string `json:"email"`
		PasswordPlainText string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.PasswordPlainText)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return

	}
	match, err := user.Password.Matches(input.PasswordPlainText)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialResponse(w, r)
		return
	}
	fmt.Println(user)
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
