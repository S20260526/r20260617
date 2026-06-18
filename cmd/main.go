package main

import (
	"app"
	"infra"
	"net/http"
)

func main() {
	h := infra.NewHandler(app.NewService())

	s := http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	e := s.ListenAndServe()

	if e != nil {
		panic(e)
	}
}
