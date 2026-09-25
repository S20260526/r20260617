package main

import (
	"context"
	"internal/app"
	"internal/coprocess"
	"internal/infra"

	"log"
	"os"
	"time"
)

type ipc struct {
	f   *os.File
	ipc *infra.GrpcIpcClient
}

func (i *ipc) SocketName() string {
	return i.f.Name()
}

func newipc() *ipc {
	f, err := os.CreateTemp(os.TempDir(), "grpcipc.*")

	if err != nil {
		log.Fatal("socket name allocate failed:", err)
	}

	defer f.Close()

	return &ipc{f, infra.NewGrpcIpcClient(f.Name())}
}

func (i *ipc) Call(ctx context.Context, rqst *app.CoprocessRequest) (*app.CoprocessResponse, error) {
	rsps, err := i.ipc.Process(ctx, rqst)

	if err != nil {
		return nil, err
	}

	return rsps, nil
}

func main() {
	payload := []byte{}

	switch len(os.Args) {
	case 0:
	case 1:
	default:
		fallthrough
	case 2:
		payload = []byte(os.Args[1])
	}

	ctx, cancel := context.WithCancel(context.Background())

	cp, err := coprocess.NewPython3(ctx, "testbench/grpcipc/main.py", newipc())

	if err != nil {
		log.Fatal("FATAL:", err)
	}

	for {
		ctx, c := context.WithTimeout(context.Background(), time.Second)

		defer c()

		rslt, err := cp.Call(ctx, &app.CoprocessRequest{Payload: payload})

		if err != nil {
			log.Println("ERR:", err)
		} else {
			log.Println("RSLT:", rslt.GetResult().String())

			cancel()

			break
		}

		time.Sleep(time.Second)
	}

	cp.Wait()
}
