package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Moji00f/GopherSocial/docs"
	"github.com/Moji00f/GopherSocial/internal/auth"
	"github.com/Moji00f/GopherSocial/internal/mailer"
	"github.com/Moji00f/GopherSocial/internal/ratelimiter"
	"github.com/Moji00f/GopherSocial/internal/store"
	"github.com/Moji00f/GopherSocial/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	version     string
	apiURL      string
	mail        mailConfig
	frontendUrl string
	auth        authConfig
	redisCfg    redisConfig
	rateLimiter ratelimiter.Config
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type basicConfig struct {
	user string
	pass string
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	mailTrap  mailTrapConfig
	sendGrid  sendGridConfig
	gmail     gmailConfig
}

type mailTrapConfig struct {
	apiKey string
}

type sendGridConfig struct {
	apiKey string
}

type gmailConfig struct {
	password string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type redisConfig struct {
	addr     string
	password string
	db       int
	enable   bool
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if app.config.rateLimiter.Enable {
		r.Use(app.RateLimiterMiddleware)
	}

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		//r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)
			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)
				r.Get("/", app.getPostHandler)

				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))

			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/active/{token}", app.activeUserHandler)

			r.Route("/{userId}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				//r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})

		})

		//publick
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)

		})

	})

	return r
}

//func (app *application) run(mux http.Handler) error {
//	//Docs
//	docs.SwaggerInfo.Version = version
//	docs.SwaggerInfo.Host = app.config.apiURL
//	docs.SwaggerInfo.BasePath = "/v1"
//	srv := &http.Server{
//		Addr:         app.config.addr,
//		Handler:      mux,
//		WriteTimeout: time.Second * 30,
//		ReadTimeout:  time.Second * 10,
//		IdleTimeout:  time.Minute,
//	}
//
//	//log.Printf("server has started at %s", app.config.addr)
//	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)
//
//	return srv.ListenAndServe()
//}

func (app *application) run(mux http.Handler) error {
	//Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()

	//log.Printf("server has started at %s", app.config.addr)
	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
