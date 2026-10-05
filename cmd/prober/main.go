package main

import (
	"bytes"
	"flag"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"
)

func main() {
	var rawURL string
	var interval time.Duration
	var size int

	flag.StringVar(&rawURL, "url", "http://localhost:8089", "URL")
	flag.DurationVar(&interval, "interval", time.Second, "duration")
	flag.IntVar(&size, "size", 1024, "payload size")

	flag.Parse()

	url, err := url.Parse(rawURL)

	if err != nil {
		log.Fatal(err)
	}

	if url.Scheme != "http" {
		log.Fatal("HTTP scheme expected")
	}

	var seed [32]byte

	copy(seed[:], []byte("0123456789ABCDEF0123456789ABCDEF"))

	rng := rand.NewChaCha8(seed)

	buf := make([]byte, size)

	tickChan := time.Tick(interval)

	for {
		rng.Read(buf)

		rsp, err := http.Post(
			url.String(), "application/octet-stream",
			bytes.NewReader(buf),
		)

		if err != nil {
			log.Print(err)
		} else {
			defer rsp.Body.Close()

			log.Print(rsp.Status)
		}

		select {
		case _, ok := <-tickChan:
			if !ok {
				log.Fatal("EXIT")
			}
		}
	}
}
