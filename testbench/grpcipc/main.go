package main

import (
	"context"
	"internal/infra"
	"os"
	"os/exec"
	"time"

	"log"
)

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

	rqst := &infra.GrpcIpcRequest{
		Payload: payload,
	}

	tmpfile, err := os.CreateTemp(os.TempDir(), "grpcipc.*")

	if err != nil {
		log.Fatal("socket name allocate failed:", err)
	}

	defer tmpfile.Close()

	socketname := tmpfile.Name()

	ipc := infra.NewGrpcIpcClient(socketname)

	cmd := exec.Command(
		"python3",
		"testbench/grpcipc/main.py", socketname,
	)

	err = cmd.Start()

	if err != nil {
		log.Fatal("cmd start failed:", err)
	}

	defer cmd.Wait()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		rsps, err := ipc.Process(ctx, rqst)

		if err != nil {
			log.Println("ERR:", err)
		} else {
			log.Println("RSP:", rsps)
		}

		cancel()

		time.Sleep(time.Second)
	}
}
