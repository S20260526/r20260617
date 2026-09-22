package main

import (
	"context"
	"internal/infra"
	"internal/msgqueue"
	"log"

	"os"
	"time"
)

func usage() {
	log.Println("usage: rabbitmq URL [push|pull]")
	os.Exit(1)
}

func goPush(url string) {
	p := msgqueue.NewPusher(url, infra.NewRMQPush("world"))

	ctx := context.Background()

	var i uint8 = 0

	for {
		log.Println("PUSH", i)

		p.Charge([]byte{i})

		for err := p.Push(ctx); err != nil; err = p.Push(ctx) {
			log.Println("ERR:", err)

			time.Sleep(time.Second)
		}

		i++

		time.Sleep(time.Second)
	}
}

func goPull(url string) {
	p := msgqueue.NewPuller([]string{url}, infra.NewRMQPull("world"))

	ctx := context.Background()

	for {
		d, err := p.Pull(ctx)

		if err != nil {
			log.Println("ERR:", err)

			time.Sleep(time.Second)
		} else {
			log.Println("PULL:", d)
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	url := os.Args[1]
	mode := os.Args[2]

	switch mode {
	default:
		usage()
	case "push":
		goPush(url)
	case "pull":
		goPull(url)
	}
}
