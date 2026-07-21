package main

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"infra"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/golang-jwt/jwt/v5"
)

type JwtStaff struct {
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

type TokenPayload struct {
	jwt.RegisteredClaims
}

func (jws *JwtStaff) NewToken() (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Claims = &TokenPayload{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}

	return t.SignedString(jws.privKey)
}

func (jws *JwtStaff) VerifyToken(t string) error {
	p, e := jwt.Parse(t, func(_ *jwt.Token) (any, error) { return jws.pubKey, nil })

	if e == nil {
		if !p.Valid {
			e = errors.New("token verification failed")
		}
	}

	return e
}

type EntryHandler struct {
	transport                 *http.Transport
	jwtStaff                  *JwtStaff
	requestCount, hcheckCount prometheus.Counter
	promHttpHandler           http.Handler
}

func panicIf(e error) {
	if e != nil {
		panic(e)
	}
}

func openJwtStaff() *JwtStaff {
	jws := &JwtStaff{}

	var e error

	keyBytes, e := os.ReadFile("jwt.key")
	panicIf(e)

	jws.privKey, e = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	panicIf(e)

	pubBytes, e := os.ReadFile("jwt.pub")
	panicIf(e)

	jws.pubKey, e = jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	panicIf(e)

	return jws
}

func openTlsConfig() *tls.Config {
	tc := &tls.Config{}

	var e error

	tc.Certificates, e = infra.ServerCertFromFile("client.crt", "client.key")
	panicIf(e)

	tc.ClientCAs, e = infra.CaCertFromFile("ca.crt")
	panicIf(e)

	tc.RootCAs, e = infra.CaCertFromFile("ca.crt")
	panicIf(e)

	return tc
}

func (h *EntryHandler) relay(tag, url string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		cl := http.Client{
			Transport: h.transport,
			Timeout:   5 * time.Second,
		}

		r, e := cl.Get(url)

		if e != nil {
			slog.Info(tag, "error", e)
			return
		}

		defer r.Body.Close()

		if r.StatusCode != 200 {
			slog.Info(tag, "status", r.StatusCode)
			return
		}

		b, e := io.ReadAll(r.Body)

		if e != nil {
			slog.Info(tag, "read error", e)
			return
		}

		ch <- string(b)
	}()

	return ch
}

func (h *EntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/hcheck" {
		h.hcheckCount.Inc()
	} else if r.RequestURI == "/prometrics" {
		h.promHttpHandler.ServeHTTP(w, r)
	} else {
		h.requestHandler(w, r)
	}
}

func (h *EntryHandler) requestHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("request", "from", r.RemoteAddr, "to", r.Host, "URI", r.RequestURI)

	h.requestCount.Inc()

	t := r.Header.Get("X-Token")

	if t == "" {
		h.noToken(w, r)
	} else if !h.verifyToken(t) {
		w.WriteHeader(http.StatusForbidden)
	} else {
		h.tokenAccepted(w, r)
	}
}

func (h *EntryHandler) noToken(w http.ResponseWriter, r *http.Request) {
	t, e := h.jwtStaff.NewToken()

	if e != nil {
		slog.Info("JWT", "token", e)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Info("JWT", "token", "created")

	w.Write([]byte(t))
}

func (h *EntryHandler) verifyToken(t string) bool {
	e := h.jwtStaff.VerifyToken(t)

	if e != nil {
		slog.Info("JWT", "token", e)

		return false
	}

	return true
}

func (h *EntryHandler) tokenAccepted(w http.ResponseWriter, r *http.Request) {
	helloCh := h.relay("hello", "https://hello:8080")
	worldCh := h.relay("world", "https://world:8080")

	hello := <-helloCh
	world := <-worldCh

	if hello == "" || world == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("%s, %s!", hello, world)))
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	pr := prometheus.NewRegistry()

	pr.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	h := EntryHandler{
		transport: &http.Transport{
			TLSClientConfig: openTlsConfig(),
		},
		jwtStaff: openJwtStaff(),
		requestCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "request_cnt",
				Help: "Hello, world! request counter",
			},
		),
		hcheckCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "hcheck_cnt",
				Help: "Healthcheck request counter",
			},
		),
		promHttpHandler: promhttp.HandlerFor(pr, promhttp.HandlerOpts{Registry: pr}),
	}

	pr.MustRegister(h.requestCount)
	pr.MustRegister(h.hcheckCount)

	s := http.Server{
		Addr:    ":8080",
		Handler: &h,
	}

	slog.Info("server", "state", "started")

	panicIf(s.ListenAndServe())
}
