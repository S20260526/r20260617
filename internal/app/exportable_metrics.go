package app

import (
	"net/http"
)

type ExportableMetrics interface {
	GetHttpHandler() http.Handler
}
