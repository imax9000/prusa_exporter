package udp

import (
	"sync"

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

	// registryMetrics = map[string]*prometheus.GaugeVec{
	// 	"last_push": lastPush,
	// }

	registryMetrics = safeRegistryMetrics{
		mu:      sync.Mutex{},
		metrics: make(map[string]*prometheus.GaugeVec),
	}
)

type safeRegistryMetrics struct {
	mu      sync.Mutex
	metrics map[string]*prometheus.GaugeVec
}

// Init initializes the Prometheus udp registry.
func Init(udpMainRegistry *prometheus.Registry) {
	udpRegistry = udpMainRegistry

	udpRegistry.MustRegister(lastPush)
	registryMetrics.mu.Lock()
	registryMetrics.metrics = make(map[string]*prometheus.GaugeVec)
	registryMetrics.metrics["last_push"] = lastPush
	registryMetrics.mu.Unlock()
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

		// Register the metric with the udp registry
		if err := udpRegistry.Register(metric); err != nil {
			log.Trace().Msgf("Metric already registered %s: %v", point.Measurement+"_"+key, err)
		}
		registryMetrics.mu.Lock()
		if existingMetric, exists := registryMetrics.metrics[point.Measurement+"_"+key]; exists {
			metric = existingMetric
		} else {
			registryMetrics.metrics[point.Measurement+"_"+key] = metric
		}
		registryMetrics.mu.Unlock()
		metric.WithLabelValues(point.Tags["mac"], point.Tags["ip"]).Set(toFloat64(value))

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
	case string:
		if v == "PLA" {
			return 1.0
		} else if v == "ABS" {
			return 2.0
		} else if v == "PETG" {
			return 3.0
		} else if v == "ASA" {
			return 4.0
		} else if v == "TPU" {
			return 5.0
		} else if v == "PC" {
			return 6.0
		} else if v == "NYLON" {
			return 7.0
		} else if v == "PVA" {
			return 8.0
		} else if v == "HIPS" {
			return 9.0
		} else if v == "PP" {
			return 10.0
		} else if v == "POM" {
			return 11.0
		} else {
			return 0.0
		}
	default:
		log.Warn().Msgf("Unsupported type %T for value %v", value, value)
		return 0.0
	}
}
