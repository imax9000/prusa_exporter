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

	registryMetrics = map[string]*prometheus.GaugeVec{
		"last_push": lastPush,
	}
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
		fmt.Println(value)
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: point.Measurement + "_" + key,
				Help: "Metric for " + point.Measurement,
			},
			[]string{"mac", "ip"},
		)

		// Register the metric with the udp registry
		if err := udpRegistry.Register(metric); err != nil {
			log.Error().Msg(fmt.Sprintf("Error registering metric %s: %v", point.Measurement+"_"+key, err))
			// If the metric is already registered, we can just use it
			if existingMetric, exists := registryMetrics[point.Measurement+"_"+key]; exists {
				metric = existingMetric
				log.Debug().Msgf("Metric %s already registered, using existing one", point.Measurement+"_"+key)
				// Set the value for the metric
				metric.WithLabelValues(point.Tags["mac"], point.Tags["ip"]).Set(toFloat64(value))
			} else {
				log.Debug().Msgf("Registering new metric %s", point.Measurement+"_"+key)
				registryMetrics[point.Measurement+"_"+key] = metric
				metric.WithLabelValues(point.Tags["mac"], point.Tags["ip"]).Set(toFloat64(value))
			}
			continue
		} else {
			if existingMetric, exists := registryMetrics[point.Measurement+"_"+key]; exists {
				metric = existingMetric
				log.Debug().Msgf("Metric %s already registered, using existing one", point.Measurement+"_"+key)
				// Set the value for the metric
				metric.WithLabelValues(point.Tags["mac"], point.Tags["ip"]).Set(toFloat64(value))
			} else {
				log.Debug().Msgf("Registering new metric %s", point.Measurement+"_"+key)
				registryMetrics[point.Measurement+"_"+key] = metric
				metric.WithLabelValues(point.Tags["mac"], point.Tags["ip"]).Set(toFloat64(value))
			}
		}

	}
}

func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case nil:
		log.Warn().Msg("Received nil value, returning 0.0")
		return 0.0
	default:
		log.Warn().Msgf("Unsupported type %T for value %v", value, value)
		return 0.0
	}
}
