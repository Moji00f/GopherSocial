package main

import (
	"net/http"
	"time"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	time.Sleep(time.Second * 5)

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {

		app.internalServerError(w, r, err)
	}

}
