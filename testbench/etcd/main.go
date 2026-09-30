package main

import (
	"context"
	"internal/app"
	"internal/infra"
	"log"
)

func main() {
	var w app.ConfigurationWatcher

	w, err := infra.NewEtcd("http://127.0.0.1:2379")

	if err != nil {
		log.Fatal("ERR:", err)
	}

	cfg, wchan, err := w.Watch(context.Background())

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
