package main

import (
	"github.com/Moji00f/GopherSocial/internal/db"
	"github.com/Moji00f/GopherSocial/internal/env"
	"github.com/Moji00f/GopherSocial/internal/mailer"
	"github.com/Moji00f/GopherSocial/internal/store"
	"go.uber.org/zap"
	"time"
)

var version = "0.0.1"

//	@title			Swagger GopherSocial
//	@description	API for GopherSocial
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securitydefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendUrl: env.GetString("FRONTEND_URL", " http://localhost:3000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/gophersocial?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:     env.GetString("ENV", "development"),
		version: version,
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_GEMAIL", ""),
			gmail: gmailConfig{
				password: env.GetString("GMAIL_PASSWORD", ""),
			},
			//fromEmail: env.GetString("FROM_EMAIL", "no-reply@example.com"),
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_API_KEY", ""),
			},
		},
	}

	//Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	//Database
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Info("Database connection pool established...")

	store := store.NewStorage(db)
	gmail, err := mailer.NewGmailClient(cfg.mail.fromEmail, cfg.mail.gmail.password)
	//mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apiKey, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: &gmail,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
