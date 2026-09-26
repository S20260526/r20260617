package main

import (
	"context"
	"internal/app"
	"internal/infra"
	"log"
)

func main() {
	var iut app.Configurator = infra.NewEtcd("http://127.0.0.1:2379")

	ctx := context.Background()

	cfg, wchan, err := iut.Watch(ctx)

	if err != nil {
		log.Fatal("ERR:", err)
	}

	log.Printf("CFG: %#v", cfg)

	for {
		select {
		case cfg := <-wchan:
			log.Printf("CFG: %#v", cfg)
		}
	}
}
