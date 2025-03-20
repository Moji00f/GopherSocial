package main

import (
	"github.com/Moji00f/GopherSocial/internal/store"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user, err := app.store.User.GetById(ctx, userId)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}

}

//func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
//
//}
