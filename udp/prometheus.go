package udp

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var (
	lastPush = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prusa_last_push_timestamp",
			Help: "Last time the printer pushed metrics to the exporter.",
		},
		[]string{"mac", "ip"},
	)
	udpRegistry *prometheus.Registry
)

// Init initializes the Prometheus udp registry.
func Init(udpMainRegistry *prometheus.Registry) {
	udpRegistry = udpMainRegistry

	udpRegistry.MustRegister(lastPush)
}

func registerMetric(point point) {
	var metric *prometheus.GaugeVec

	for key, value := range point.Fields {
		// Create a new metric with the given point

		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: point.Measurement + "_" + key,
				Help: "Metric for " + point.Measurement,
			},
			[]string{"mac", "ip"},
		)

		metric.WithLabelValues(point.Tags["mac"], point.Tags["ip"]).Set(toFloat64(value))

	}

	// Register the metric with the registry
	if err := udpRegistry.Register(metric); err != nil {
		log.Error().Msg(fmt.Sprintf("Error registering metric %s: %v", point.Measurement, err))
	}
}

func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	default:
		log.Debug().Msg(fmt.Sprintf("Unsupported type for conversion to float64: %T", value))
		return 0.0
	}
}
