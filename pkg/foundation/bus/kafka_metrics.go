package bus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	kafkaReaderMessagesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "reader",
		Name:      "messages_total",
		Help:      "Total messages read by topic (cumulative from reader stats).",
	}, []string{"topic"})

	kafkaReaderLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "reader",
		Name:      "lag",
		Help:      "Current consumer lag by topic.",
	}, []string{"topic"})

	kafkaReaderErrorsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "reader",
		Name:      "errors_total",
		Help:      "Total read errors by topic (cumulative from reader stats).",
	}, []string{"topic"})

	kafkaWriterMessagesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "writer",
		Name:      "messages_total",
		Help:      "Total messages written by topic (cumulative from writer stats).",
	}, []string{"topic"})

	kafkaWriterErrorsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "writer",
		Name:      "errors_total",
		Help:      "Total write errors by topic (cumulative from writer stats).",
	}, []string{"topic"})

	kafkaWriterBatchAvgSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "writer",
		Name:      "batch_avg_seconds",
		Help:      "Average batch write duration in seconds by topic.",
	}, []string{"topic"})
)

// StartStatsCollector starts a goroutine that periodically collects reader/writer
// stats and updates Prometheus gauges. It stops when ctx is cancelled.
func (c *kafkaClient) StartStatsCollector(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collectStats()
		}
	}
}

func (c *kafkaClient) collectStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for topic, reader := range c.readers {
		stats := reader.Stats()
		kafkaReaderMessagesTotal.WithLabelValues(topic).Set(float64(stats.Messages))
		kafkaReaderLag.WithLabelValues(topic).Set(float64(stats.Lag))
		kafkaReaderErrorsTotal.WithLabelValues(topic).Set(float64(stats.Errors))
	}

	for topic, writer := range c.writers {
		stats := writer.Stats()
		kafkaWriterMessagesTotal.WithLabelValues(topic).Set(float64(stats.Messages))
		kafkaWriterErrorsTotal.WithLabelValues(topic).Set(float64(stats.Errors))
		kafkaWriterBatchAvgSeconds.WithLabelValues(topic).Set(stats.BatchTime.Avg.Seconds())
	}
}
