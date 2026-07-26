package main

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"infra"

	_ "m20260618-entry/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/golang-jwt/jwt/v5"
)

const HelloHostPort = "hello:8081"
const WorldHostPort = "world:8080"

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
	statusHtml      *html.Template
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

func (h *HelloHandler) dial(hostPort string) error {
	c, e := tls.Dial("tcp", hostPort, h.clientTransport.TLSClientConfig)

	if e == nil {
		defer c.Close()
	} else {
		slog.Info("dial", "TLS error", e)
	}

	return e
}

func (h *HelloHandler) relay(tag, hostPort string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		cl := http.Client{
			Transport: h.clientTransport,
			Timeout:   5 * time.Second,
		}

		r, e := cl.Get("https://" + hostPort)

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

func (h *HelloHandler) Status(c echo.Context) error {
	hl := "ok"
	w := "ok"

	if h.dial(HelloHostPort) != nil {
		hl = "err"
	}

	if h.dial(WorldHostPort) != nil {
		w = "err"
	}

	e := h.statusHtml.Execute(
		c.Response().Writer,
		struct {
			Hello, World string
		}{
			Hello: hl,
			World: w,
		},
	)

	if e != nil {
		slog.Info("status", "error", e)
	}

	return nil
}

// @Summary		greeting
// @Description	return simple greeting
// @Tags			hello
// @Produce		text/plain
// @Param			X-Token	header		string	false	"JWT access token"
// @Success		200		{string}	string
// @Failure		403
// @Failure		500
// @Router			/ [get]
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
	helloCh := h.relay("hello", HelloHostPort)
	worldCh := h.relay("world", WorldHostPort)

	hello := <-helloCh
	world := <-worldCh

	if hello == "" || world == "" {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.String(http.StatusOK, fmt.Sprintf("%s, %s!", hello, world))
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
		statusHtml: html.Must(
			html.New("status").Parse(`
			  {{ define "statusBlock" }}
			    {{ if eq . "ok" }}
			      <p style="color: green;">Online</p>
			    {{ else if eq . "err" }}
			      <p style="color: red;">Offline</p>
			    {{ else }}
			      <p>N/A</p>
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
				<h2>World:</h2>
				{{ template "statusBlock" .World }}
			      </body>
			    </html>
			`),
		),
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

	e.GET("/prometrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/status", h.Status)

	e.GET("/", h.Handle)

	slog.Info("server", "state", "starting")

	panicIf(e.Start(":8080"))
}
