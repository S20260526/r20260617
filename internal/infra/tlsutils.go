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

	defer f.Close()

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

func LoadTlsConfig() (*tls.Config, error) {
	tc := &tls.Config{}

	var e error

	tc.Certificates, e = ServerCertFromFile("server.crt", "server.key")

	if e != nil {
		return tc, e
	}

	tc.ClientCAs, e = CaCertFromFile("ca.crt")

	if e != nil {
		return tc, e
	}

	tc.ClientAuth = tls.RequireAndVerifyClientCert

	return tc, nil
}
