package pgxpoolprometheus

/**
 * (C) Copyright IBM Corp. 2021.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector is a prometheus.Collector that will collect the nine statistics produced by pgxpool.Stat.
type Collector struct {
	stater stater

	acquireCountDesc         *prometheus.Desc
	acquireDurationDesc      *prometheus.Desc
	acquiredConnsDesc        *prometheus.Desc
	canceledAcquireCountDesc *prometheus.Desc
	constructingConnsDesc    *prometheus.Desc
	emptyAcquireCountDesc    *prometheus.Desc
	idleConnsDesc            *prometheus.Desc
	maxConnsDesc             *prometheus.Desc
	totalConnsDesc           *prometheus.Desc
}

// Stater is a provider of the Stat() function. Implemented by pgxpool.Pool.
type Stater interface {
	Stat() *pgxpool.Stat
}

// NewCollector creates a new Collector to collect stats from pgxpool.
func NewCollector(stater Stater, labels map[string]string) *Collector {
	return newCollector(&staterWrapper{stater}, labels)
}

// newCollector is an internal only constructor for a Collect that uses a wrapped Stater or other provider of stats.
// labels are provided as prometheus.Labels to each metric and may be nil. A label is recommended when an application uses more than one pgxpool.Pool to enable differentiation between them.
func newCollector(stater stater, labels map[string]string) *Collector {
	return &Collector{
		stater: stater,
		acquireCountDesc: prometheus.NewDesc(
			"pgxpool_acquire_count",
			"Cumulative count of successful acquires from the pool.",
			nil, labels),
		acquireDurationDesc: prometheus.NewDesc(
			"pgxpool_acquire_duration_ns",
			"Total duration of all successful acquires from the pool in nanoseconds.",
			nil, labels),
		acquiredConnsDesc: prometheus.NewDesc(
			"pgxpool_acquired_conns",
			"Number of currently acquired connections in the pool.",
			nil, labels),
		canceledAcquireCountDesc: prometheus.NewDesc(
			"pgxpool_canceled_acquire_count",
			"Cumulative count of acquires from the pool that were canceled by a context.",
			nil, labels),
		constructingConnsDesc: prometheus.NewDesc(
			"pgxpool_constructing_conns",
			"Number of conns with construction in progress in the pool.",
			nil, labels),
		emptyAcquireCountDesc: prometheus.NewDesc(
			"pgxpool_empty_acquire",
			"Cumulative count of successful acquires from the pool that waited for a resource to be released or constructed because the pool was empty.",
			nil, labels),
		idleConnsDesc: prometheus.NewDesc(
			"pgxpool_idle_conns",
			"Number of currently idle conns in the pool.",
			nil, labels),
		maxConnsDesc: prometheus.NewDesc(
			"pgxpool_max_conns",
			"Maximum size of the pool.",
			nil, labels),
		totalConnsDesc: prometheus.NewDesc(
			"pgxpool_total_conns",
			"Total number of resources currently in the pool. The value is the sum of ConstructingConns, AcquiredConns, and IdleConns.",
			nil, labels),
	}
}

// Describe implements the prometheus.Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	stats := c.stater.stat()
	metrics <- prometheus.MustNewConstMetric(
		c.acquireCountDesc,
		prometheus.CounterValue,
		stats.acquireCount(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.acquireDurationDesc,
		prometheus.CounterValue,
		stats.acquireDuration(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.acquiredConnsDesc,
		prometheus.GaugeValue,
		stats.acquiredConns(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.canceledAcquireCountDesc,
		prometheus.CounterValue,
		stats.canceledAcquireCount(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.constructingConnsDesc,
		prometheus.GaugeValue,
		stats.constructingConns(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.emptyAcquireCountDesc,
		prometheus.CounterValue,
		stats.emptyAcquireCount(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.idleConnsDesc,
		prometheus.GaugeValue,
		stats.idleConns(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxConnsDesc,
		prometheus.GaugeValue,
		stats.maxConns(),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.totalConnsDesc,
		prometheus.GaugeValue,
		stats.totalConns(),
	)
}

// stater is an internal only version of Stater making it easier to mock or use alternatives to the pgxpool.Stat struct.
type stater interface {
	stat() stat
}

// stat is an interface version of the pgxpool.Stat struct which also returns float64 values for ease of use with Prometheus.
type stat interface {
	acquireCount() float64
	acquireDuration() float64
	acquiredConns() float64
	canceledAcquireCount() float64
	constructingConns() float64
	emptyAcquireCount() float64
	idleConns() float64
	maxConns() float64
	totalConns() float64
}

// staterWrapper converts a Stater into a stater and is the concrete implementation of the stater interface that uses a real pgxpool.Pool.
type staterWrapper struct {
	stater Stater
}

func (w *staterWrapper) stat() stat {
	return &statWrapper{w.stater.Stat()}
}

type statWrapper struct {
	stats *pgxpool.Stat
}

func (w *statWrapper) acquireCount() float64 {
	return float64(w.stats.AcquireCount())
}
func (w *statWrapper) acquireDuration() float64 {
	return float64(w.stats.AcquireDuration())
}
func (w *statWrapper) acquiredConns() float64 {
	return float64(w.stats.AcquiredConns())
}
func (w *statWrapper) canceledAcquireCount() float64 {
	return float64(w.stats.CanceledAcquireCount())
}
func (w *statWrapper) constructingConns() float64 {
	return float64(w.stats.ConstructingConns())
}
func (w *statWrapper) emptyAcquireCount() float64 {
	return float64(w.stats.EmptyAcquireCount())
}
func (w *statWrapper) idleConns() float64 {
	return float64(w.stats.IdleConns())
}
func (w *statWrapper) maxConns() float64 {
	return float64(w.stats.MaxConns())
}
func (w *statWrapper) totalConns() float64 {
	return float64(w.stats.TotalConns())
}