package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler exposes Prometheus metrics at /metrics
func Handler() http.Handler {
	return promhttp.Handler()
}
