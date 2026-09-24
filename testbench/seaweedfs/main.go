package main

import (
	"context"
	"internal/infra"
	"log"
	"time"
)

func main() {
	swfs := infra.NewSeaWeedFS("localhost:9333")

	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*5),
	)

	defer cancel()

	id, err := swfs.Create(ctx, []byte("1234"))

	if err != nil {
		log.Fatal("ERR:", err)
	} else {
		log.Println("ID:", id)
	}

	data, err := swfs.Read(ctx, id)

	if err != nil {
		log.Fatal("ERR:", err)
	} else {
		log.Println("RDBACK:", string(data))
	}

	err = swfs.Delete(ctx, id)

	if err != nil {
		log.Fatal("ERR:", err)
	} else {
		log.Println("DEL")
	}

	_, err = swfs.Read(ctx, id)

	if err != nil {
		log.Println("DEL:", err)
	} else {
		log.Fatal("NOT DELETED")
	}
}
