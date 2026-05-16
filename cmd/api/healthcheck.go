package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	env := envelope{
		"status": "available",
		"SystemInfo": map[string]string{

			"environment": app.config.env,
			"version":     version,
		},
	}
	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
