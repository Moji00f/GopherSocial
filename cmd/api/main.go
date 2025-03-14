package main

import (
	"github.com/Moji00f/GopherSocial/internal/env"
	"log"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	app := &application{
		config: cfg,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
