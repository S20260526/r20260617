package main

import (
	"context"
	"internal/grpcipc"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log"
)

func main() {
	socketpath := "/tmp/grpcipc.socket"
	payload := []byte{}

	switch len(os.Args) {
	case 0:
	case 1:
	default:
		fallthrough
	case 3:
		payload = []byte(os.Args[2])
		fallthrough
	case 2:
		socketpath = os.Args[1]
	}

	conn, err := grpc.NewClient(
		"unix://"+socketpath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("ERR:", err)
	}

	ipc := grpcipc.NewIpcClient(conn)

	rqst := &grpcipc.Request{
		Payload: payload,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	rsps, err := ipc.Process(ctx, rqst)

	if err != nil {
		log.Println("ERR:", err)
	} else {
		log.Println("RSP:", rsps)
	}

	cancel()
}
