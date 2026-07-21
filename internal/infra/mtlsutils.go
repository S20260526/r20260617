package infra

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"os"
)

func CaCertFromFile(certFile string) (*x509.CertPool, error) {
	f, e := os.Open(certFile)

	if e != nil {
		return nil, e
	}

	b, e := io.ReadAll(f)

	if e != nil {
		return nil, e
	}

	p := x509.NewCertPool()

	if !p.AppendCertsFromPEM(b) {
		return nil, errors.New("PEM parse failed")
	}

	return p, nil
}

func ServerCertFromFile(certFile, keyFile string) ([]tls.Certificate, error) {
	c, e := tls.LoadX509KeyPair(certFile, keyFile)

	if e != nil {
		return []tls.Certificate{}, e
	}

	return []tls.Certificate{c}, nil
}
