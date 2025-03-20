package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Moji00f/GopherSocial/internal/store"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type userKey string

var userCtx userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserFromContext(r)
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}

}

type FollowUser struct {
	UserId int64 `json:"user_id"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)

	//TODO: revert back to auth from ctx
	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestRequest(w, r, err)
		return
	}
	fmt.Println(payload.UserId, followerUser.ID)

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followerUser.ID, payload.UserId); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unFollowerUser := getUserFromContext(r)
	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}

	//TODO: revert back to auth from ctx
	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestRequest(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.UnFollow(ctx, unFollowerUser.ID, payload.UserId); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
		if err != nil {
			app.badRequestRequest(w, r, err)
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

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user := r.Context().Value(userCtx).(*store.User)
	return user
}
