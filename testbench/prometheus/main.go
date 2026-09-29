package main

import (
	"internal/app"
	"internal/infra"
	"net/http"
	"time"
)

func main() {
	p := infra.NewPrometrics()

	var m app.WorkerMetrics = p

	http.Handle("/metrics", p.GetHttpHandler())
	http.HandleFunc(
		"/yes",
		func(w http.ResponseWriter, r *http.Request) {
			m.RegIn()
			m.RegYes()
			w.Write([]byte("YES\n"))
		},
	)
	http.HandleFunc(
		"/no",
		func(w http.ResponseWriter, r *http.Request) {
			m.RegIn()
			m.RegNo()
			w.Write([]byte("NO\n"))
		},
	)
	http.HandleFunc(
		"/err",
		func(w http.ResponseWriter, r *http.Request) {
			m.RegIn()
			m.RegErr()
			w.Write([]byte("ERR\n"))
		},
	)
	http.HandleFunc(
		"/fail",
		func(w http.ResponseWriter, r *http.Request) {
			m.RegIn()
			m.RegFail()
			w.Write([]byte("FAIL\n"))
		},
	)
	http.HandleFunc(
		"/out",
		func(w http.ResponseWriter, r *http.Request) {
			m.RegIn()
			p.RegOut()
			w.Write([]byte("OUT\n"))
		},
	)
	http.HandleFunc(
		"/time",
		func(w http.ResponseWriter, r *http.Request) {
			m.RegIn()
			t := time.Now()

			time.Sleep(time.Second)

			m.RegFrontToEndDuration(t)

			w.Write([]byte("TIME\n"))
		},
	)

	http.ListenAndServe(":8080", nil)
}
