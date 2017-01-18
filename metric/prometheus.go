package metric

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var promMu sync.Mutex

func promRegister(c prometheus.Collector) prometheus.Collector {
	promMu.Lock()
	defer promMu.Unlock()
	if err := prometheus.Register(c); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return are.ExistingCollector
		}
		panic(err)
	}
	return c
}

type prom struct {
	lbls map[string]string
}

func Prometheus() (Node, http.Handler) {
	return &prom{lbls: make(map[string]string)}, prometheus.Handler()
}

func (node *prom) opts(name, help string) prometheus.Opts {
	return prometheus.Opts{
		Name:        name,
		Help:        help,
		ConstLabels: node.lbls,
	}
}

func (node *prom) Counter(name, desc string, labels ...string) Counter {
	opts := node.opts(name, desc)
	hopts := prometheus.CounterOpts{
		Namespace:   opts.Namespace,
		Subsystem:   opts.Subsystem,
		Name:        opts.Name,
		Help:        opts.Help,
		ConstLabels: opts.ConstLabels,
	}
	h := promRegister(prometheus.NewCounterVec(hopts, labels)).(*prometheus.CounterVec)
	return func(v float64, labels ...string) {
		h.WithLabelValues(labels...).Add(v)
	}
}

func (node *prom) Sampler(name, desc string, min, max float64, labels ...string) Sampler {
	opts := node.opts(name, desc)
	hopts := prometheus.HistogramOpts{
		Namespace:   opts.Namespace,
		Subsystem:   opts.Subsystem,
		Name:        opts.Name,
		Help:        opts.Help,
		ConstLabels: opts.ConstLabels,
		Buckets:     (bktRange{min: min, max: max}).exponentialBuckets(),
	}
	h := promRegister(prometheus.NewHistogramVec(hopts, labels)).(*prometheus.HistogramVec)
	return func(v float64, labels ...string) {
		h.WithLabelValues(labels...).Observe(v)
	}
}

func (node *prom) Lbl(k, v string) Node {
	out := make(map[string]string, len(node.lbls))
	for oldk, oldv := range node.lbls {
		out[oldk] = oldv
	}
	out[k] = v
	return &prom{lbls: out}
}
