package infra

import (
	"log/slog"
	"crypto/tls"
	"net/http"
)

func StartMTLSServer(h Handler) error {
	tc := &tls.Config{}

	var e error

	tc.Certificates, e = ServerCertFromFile("server.crt", "server.key")

	if e != nil {
		return e
	}

	tc.ClientCAs, e = CaCertFromFile("ca.crt")

	if e != nil {
		return e
	}

	tc.ClientAuth = tls.RequireAndVerifyClientCert

	s := http.Server{
		Addr:      ":8080",
		Handler:   h,
		TLSConfig: tc,
	}

	slog.Info("mtls", "state", "started")

	return s.ListenAndServeTLS("", "")
}
