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

	port := Config.Get("mtlsport", "8080")

	s := http.Server{
		Addr:      ":" + port,
		Handler:   h,
		TLSConfig: tc,
	}

	slog.Info("mtls", "state", "starting")

	e = s.ListenAndServeTLS("", "")

	slog.Info("mtls", "state", "exit")

	return e
}
