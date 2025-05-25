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

// Init initializes the Prometheus udp registry.
func Init(udpRegistry *prometheus.Registry) {
	if udpRegistry == nil {
		log.Panic("UDP Registry cannot be nil")
	}

	udpRegistry.MustRegister(up)

}
