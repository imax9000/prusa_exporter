package udp

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	lastPush = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prusa_last_push_timestamp",
			Help: "Last time the printer pushed metrics to the exporter.",
		},
		[]string{"ip", "mac"},
	)
)

// Init initializes the Prometheus udp registry.
func Init(udpRegistry *prometheus.Registry) {
	if udpRegistry == nil {
		log.Panic("UDP Registry cannot be nil")
	}

	udpRegistry.MustRegister(lastPush)
}
