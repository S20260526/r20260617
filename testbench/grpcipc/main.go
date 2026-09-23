package main

import (
	"context"
	"internal/infra"
	"os"
	"time"

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

	ipc := infra.NewGrpcIpcClient(socketpath)

	rqst := &infra.GrpcIpcRequest{
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
