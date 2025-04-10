package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("internal server error: %s path: %s error:%s", r.Method, r.URL.Path, err)

	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err)
	writeJSONError(w, http.StatusInternalServerError, "The server encountered a problem")
}

func (app *application) badRequestRequest(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("bad request error: %s path: %s error:%s", r.Method, r.URL.Path, err)

	app.logger.Warnw("bad request", "method", r.Method, "path", r.URL.Path, "error", err)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("not found error: %s path: %s error:%s", r.Method, r.URL.Path, err)

	app.logger.Warnw("not found error", "method", r.Method, "path", r.URL.Path, "error", err)
	writeJSONError(w, http.StatusNotFound, "not found")
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("conflict response: %s path: %s error:%s", r.Method, r.URL.Path, err)

	app.logger.Errorw("conflict response", "method", r.Method, "path", r.URL.Path, "error", err)
	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "err", err.Error())
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSON(w, http.StatusUnauthorized, err.Error())
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized error", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSON(w, http.StatusUnauthorized, err.Error())
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)

	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
