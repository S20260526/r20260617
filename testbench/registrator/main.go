package main

import (
	"context"
	"internal/app"
	"internal/infra"
	"log"
	"time"
)

func main() {
	var r app.Registrator

	ctx := context.Background()

	r, err := infra.NewRDBMS(
		ctx, "postgres",
		"host=localhost dbname=testdb sslmode=disable user=postgres password=1234",
	)

	if err != nil {
		log.Fatal("ERR:", err)
	}

	err = r.Put(ctx, "events", app.Event{T: time.Now(), Id: "1234"})

	if err != nil {
		log.Fatal("ERR:", err)
	} else {
		log.Println("OK")
	}
}
