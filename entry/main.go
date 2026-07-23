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

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

type HelloHandler struct {
	clientTransport *http.Transport
	jwtStaff        *JwtStaff
	requestCount    prometheus.Counter
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

func (h *HelloHandler) relay(tag, url string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		cl := http.Client{
			Transport: h.clientTransport,
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

func (h *HelloHandler) Handle(c echo.Context) error {
	r := c.Request()

	slog.Info("request", "from", r.RemoteAddr, "to", r.Host, "URI", r.RequestURI)

	h.requestCount.Inc()

	t := r.Header.Get("X-Token")

	if t == "" {
		return h.noToken(c)
	} else if !h.verifyToken(t) {
		return c.NoContent(http.StatusForbidden)
	} else {
		return h.tokenAccepted(c)
	}
}

func (h *HelloHandler) noToken(c echo.Context) error {
	t, e := h.jwtStaff.NewToken()

	if e != nil {
		slog.Info("JWT", "token", e)

		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Info("JWT", "token", "created")

	return c.String(http.StatusOK, t)
}

func (h *HelloHandler) verifyToken(t string) bool {
	e := h.jwtStaff.VerifyToken(t)

	if e != nil {
		slog.Info("JWT", "token", e)

		return false
	}

	return true
}

func (h *HelloHandler) tokenAccepted(c echo.Context) error {
	helloCh := h.relay("hello", "https://hello:8080")
	worldCh := h.relay("world", "https://world:8080")

	hello := <-helloCh
	world := <-worldCh

	if hello == "" || world == "" {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.String(http.StatusOK, fmt.Sprintf("%s, %s!", hello, world))
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	pr := prometheus.NewRegistry()

	pr.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	hcheckCount := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "hcheck_cnt",
			Help: "Healthcheck request counter",
		},
	)

	pr.MustRegister(hcheckCount)

	h := HelloHandler{
		clientTransport: &http.Transport{
			TLSClientConfig: openTlsConfig(),
		},
		jwtStaff: openJwtStaff(),
		requestCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "request_cnt",
				Help: "Hello, world! request counter",
			},
		),
	}

	pr.MustRegister(h.requestCount)

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())

	e.GET(
		"/hcheck", func(c echo.Context) error {
			hcheckCount.Inc()

			return c.NoContent(http.StatusOK)
		},
	)

	e.GET("/prometrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/", h.Handle)

	slog.Info("server", "state", "starting")

	panicIf(e.Start(":8080"))
}
