package main

import (
	"context"
	"fmt"
	"internal/infra"
	"internal/msgqueue"
	"log"

	"os"
	"time"
)

func usage() {
	fmt.Println("usage: rabbitmq URL [push|pull]")
	os.Exit(1)
}

func goPush(conn *infra.RMQConnection, url string) {
	p := msgqueue.NewPusher(url, infra.NewRMQPublishing(conn, "world"))

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

func goPull(conn *infra.RMQConnection, url string) {
	p := msgqueue.NewPuller([]string{url}, infra.NewRMQConsuming(conn, "world"))

	ctx := context.Background()

	for {
		d, err := p.Pull(ctx)

		if err != nil {
			log.Println("ERR:", err)

			time.Sleep(time.Second)
		} else {
			log.Println("PULL:", d.GetData())

			d.Acknowledge()
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	url := os.Args[1]
	mode := os.Args[2]

	conn := infra.NewRMQConnection()

	switch mode {
	default:
		usage()
	case "push":
		goPush(conn, url)
	case "pull":
		goPull(conn, url)
	}
}
