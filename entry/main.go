package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"infra"

	ft "fubotorp"

	_ "m20260618-entry/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/golang-jwt/jwt/v5"

	grpc "google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"

	vlkg "github.com/valkey-io/valkey-go"
	vlkl "github.com/valkey-io/valkey-go/valkeylimiter"
)

type Backend struct {
	host     string
	httpPort string
	grpcPort string

	rest http.Client
	grpc ft.ServiceClient
}

func (b *Backend) openRest(tc *tls.Config) {
	b.httpPort = infra.Config.Get("mtlsport", "8080")
	b.rest = http.Client{
		Transport: &http.Transport{TLSClientConfig: tc},
		Timeout:   5 * time.Second,
	}
}

func (b *Backend) openGrpc(tc *tls.Config) {
	b.grpcPort = infra.Config.Get("grpcport", "8081")
	c8n, e := grpc.NewClient(
		b.grpcHostPort(),
		grpc.WithTransportCredentials(grpccred.NewTLS(tc)),
	)

	panicIf(e)

	b.grpc = ft.NewServiceClient(c8n)
}

func (b *Backend) httpHostPort() string {
	if b.httpPort == "" {
		panic("HTTP port not yet set")
	}

	return fmt.Sprintf("%s:%s", b.host, b.httpPort)
}

func (b *Backend) grpcHostPort() string {
	if b.grpcPort == "" {
		panic("gRPC port not yet set")
	}

	return fmt.Sprintf("%s:%s", b.host, b.grpcPort)
}

func (b *Backend) restUrl() string {
	return fmt.Sprintf("https://%s", b.httpHostPort())
}

var HelloBackend = &Backend{host: "hello"}
var WorldBackend = &Backend{host: "world"}

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
	tlsClientConfig *tls.Config
	jwtStaff        *JwtStaff
	requestCount    prometheus.Counter
	statusHtml      *html.Template
	rateLimiter     vlkl.RateLimiterClient
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

func openRateLimiter() vlkl.RateLimiterClient {
	switch strings.ToLower(infra.Config.Get("ratelimit", "off")) {
	case "off", "no", "false", "disable", "disabled":
		return nil
	}

	lmtr, e := vlkl.NewRateLimiter(
		vlkl.RateLimiterOption{
			ClientOption: vlkg.ClientOption{
				InitAddress: []string{
					infra.Config.Get(
						"valkeyhostport", "valkey:6379",
					),
				},
			},
			KeyPrefix: "rate_limiter",
			Limit:     1,
			Window:    5 * time.Second,
		},
	)

	if e != nil {
		slog.Info("rate limiter", "create error", e)
		lmtr = nil
	}

	return lmtr
}

func (h *HelloHandler) rateLimited(ctx context.Context, key string) bool {
	if h.rateLimiter != nil {
		r, e := h.rateLimiter.Allow(ctx, key)

		if e != nil {
			slog.Info("rate limiter", "runtime error", e)
		} else {
			return !r.Allowed
		}
	}

	return false
}

func (h *HelloHandler) ping(b *Backend) error {
	c, e := tls.Dial("tcp", b.httpHostPort(), h.tlsClientConfig)

	if e == nil {
		defer c.Close()
	} else {
		slog.Info("ping", "TLS error", e)
	}

	return e
}

func (h *HelloHandler) httpRelay(b *Backend) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		cl := b.rest

		r, e := cl.Get(b.restUrl())

		if e != nil {
			slog.Info("HTTP relay", "host", b.host, "error", e)
			return
		}

		defer r.Body.Close()

		if r.StatusCode != 200 {
			slog.Info("HTTP relay", "host", b.host, "status", r.StatusCode)
			return
		}

		term, e := io.ReadAll(r.Body)

		if e != nil {
			slog.Info("HTTP relay", "host", b.host, "read error", e)
			return
		}

		ch <- string(term)
	}()

	return ch
}

func (h *HelloHandler) grpcRelay(b *Backend) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		ctx, cncl := context.WithTimeout(
			context.Background(), time.Second*5,
		)

		defer cncl()

		rsps, e := b.grpc.Get(ctx, &ft.Rqst{})

		if e != nil {
			slog.Info("gRPC relay", "host", b.host, "error", e)
			return
		}

		ch <- rsps.Message
	}()

	return ch
}

func (h *HelloHandler) Status(c echo.Context) error {
	if h.rateLimited(c.Request().Context(), "status") {
		return c.NoContent(http.StatusTooManyRequests)
	}

	hl := "ok"
	w := "ok"

	if e := h.ping(HelloBackend); e != nil {
		hl = e.Error()
	}

	if e := h.ping(WorldBackend); e != nil {
		w = e.Error()
	}

	if e := h.statusHtml.Execute(
		c.Response().Writer,
		struct {
			Hello, World string
		}{
			Hello: hl,
			World: w,
		},
	); e != nil {
		slog.Info("status", "error", e)
	}

	return nil
}

// @Summary		greeting
// @Description	return simple greeting via RESTful backend API
// @Tags			hello
// @Produce		text/plain
// @Param			X-Token	header		string	false	"JWT access token"
// @Success		200		{string}	string
// @Failure		403
// @Failure		500
// @Router			/rest [get]
func (h *HelloHandler) Rest(c echo.Context) error {
	return h.handle(c)
}

// @Summary		greeting
// @Description	return simple greeting via gRPC backend API
// @Tags			hello
// @Produce		text/plain
// @Param			X-Token	header		string	false	"JWT access token"
// @Success		200		{string}	string
// @Failure		403
// @Failure		500
// @Router			/grpc [get]
func (h *HelloHandler) Grpc(c echo.Context) error {
	return h.handle(c)
}

func (h *HelloHandler) handle(c echo.Context) error {
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
	if h.rateLimited(c.Request().Context(), "newtoken") {
		return c.NoContent(http.StatusTooManyRequests)
	}

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
	var hCh, wCh chan string

	switch c.Path() {
	case "/rest":
		hCh = h.httpRelay(HelloBackend)
		wCh = h.httpRelay(WorldBackend)
	case "/grpc":
		hCh = h.grpcRelay(HelloBackend)
		wCh = h.grpcRelay(WorldBackend)
	default:
		return c.NoContent(http.StatusNotFound)
	}

	hl := <-hCh
	wr := <-wCh

	if hl == "" || wr == "" {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.String(http.StatusOK, fmt.Sprintf("%s, %s!", hl, wr))
}

// @title			Hello world entry API
// @version		1.0
// @description	Hello world entry point
// @contact.name	A.U.Thor
// @contact.email	dont.spam.me@example.com
// @license.name	GNU
// @BasePath		/
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

	tc := openTlsConfig()

	h := HelloHandler{
		tlsClientConfig: tc,
		jwtStaff:        openJwtStaff(),
		requestCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "request_cnt",
				Help: "Hello, world! request counter",
			},
		),
		statusHtml: html.Must(
			html.New("status").Parse(`
			  {{ define "statusBlock" }}
			    {{ if eq . "ok" }}
			      <p style="color: green;">Online</p>
			    {{ else }}
			      <p style="color: red;">Offline</p>
			    {{ end }}
			  {{ end }}
			  <!DOCTYPE html>
			    <html>
			      <head>
			        <meta charset="latin-1">
			        <title>Status</title>
			      </head>
			      <body>
			        <h1>Status</h1>
				<h2>Hello:</h2>
				{{ template "statusBlock" .Hello }}
				<p>{{ .Hello }}</p>
				<h2>World:</h2>
				{{ template "statusBlock" .World }}
				<p>{{ .World }}</p>
			      </body>
			    </html>
			`),
		),
		rateLimiter: openRateLimiter(),
	}

	pr.MustRegister(h.requestCount)

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET(
		"/swagit", func(c echo.Context) error {
			return c.Redirect(
				http.StatusMovedPermanently,
				"/swagger/index.html",
			)
		},
	)

	e.GET(
		"/hcheck", func(c echo.Context) error {
			hcheckCount.Inc()

			return c.NoContent(http.StatusOK)
		},
	)

	e.GET("/prometrics", echo.WrapHandler(promhttp.HandlerFor(pr, promhttp.HandlerOpts{})))

	e.GET("/status", h.Status)

	e.GET("/rest", h.Rest)
	e.GET("/grpc", h.Grpc)

	HelloBackend.openRest(tc)
	WorldBackend.openRest(tc)

	HelloBackend.openGrpc(tc)
	WorldBackend.openGrpc(tc)

	slog.Info("server", "state", "starting")

	panicIf(e.Start(":8080"))
}
