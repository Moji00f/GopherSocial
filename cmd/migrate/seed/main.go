package main

import (
	"github.com/Moji00f/GopherSocial/internal/db"
	"github.com/Moji00f/GopherSocial/internal/env"
	"github.com/Moji00f/GopherSocial/internal/store"
	"log"
)

func main() {
	//addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/gophersocial?sslmode=disable")
	dsn := env.GetString("DB_ADDR1", "postgres://admin:adminpassword@localhost/gophersocial?sslmode=disable")

	//fmt.Println(addr)

	conn, err := db.New(dsn, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store, conn)

}
