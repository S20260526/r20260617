package infra

import (
	"log/slog"
	"net/http"
)

func StartMTLSServer(h Handler) error {
	tc, e := LoadTlsConfig()

	if e != nil {
		return e
	}

	s := http.Server{
		Addr:      ":8080",
		Handler:   h,
		TLSConfig: tc,
	}

	slog.Info("mtls", "state", "starting")

	e = s.ListenAndServeTLS("", "")

	slog.Info("mtls", "state", "exit")

	return e
}
