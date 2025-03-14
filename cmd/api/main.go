package main

import (
	"github.com/Moji00f/GopherSocial/internal/env"
	"github.com/Moji00f/GopherSocial/internal/store"
	"log"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	store := store.NewStorage(nil)
	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
