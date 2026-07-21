package main

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"

	"app"
	"infra"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	tc := &tls.Config{}

	var e error

	tc.Certificates, e = infra.ServerCertFromFile("server.crt", "server.key")

	if e != nil {
		panic(e)
	}

	tc.ClientCAs, e = infra.CaCertFromFile("ca.crt")

	if e != nil {
		panic(e)
	}

	tc.ClientAuth = tls.RequireAndVerifyClientCert

	s := http.Server{
		Addr:      ":8080",
		Handler:   infra.NewHandler(app.NewWorldService()),
		TLSConfig: tc,
	}

	slog.Info("world", "state", "started")

	e = s.ListenAndServeTLS("", "")

	if e != nil {
		panic(e)
	}
}
