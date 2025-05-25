package lineprotocol

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	up = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prusa_up",
			Help: "Returns if printer is online.",
		},
		[]string{"ip", "mac"},
	)
)

// Init initializes the Prometheus line protocol registry.
func Init(lineProtocolRegistry *prometheus.Registry) {
	if lineProtocolRegistry == nil {
		log.Panic("lineProtocolRegistry cannot be nil")
	}

	lineProtocolRegistry.MustRegister(up)

}
