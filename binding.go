package metricsopentelemetry

import (
	"context"
	"strings"

	logging "github.com/ipfs/go-log"
	metrics "github.com/ipfs/go-metrics-interface"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var log logging.EventLogger = logging.Logger("metrics-opentelemetry")

func Inject() error {
	return metrics.InjectImpl(newCreator)
}

func newCreator(name, helptext string) metrics.Creator {
	name = strings.Replace(name, ".", "_", -1)
	return &creator{
		meter:    global.Meter(name),
		name:     name,
		helptext: helptext,
	}
}

var _ metrics.Creator = &creator{}

type creator struct {
	meter    metric.Meter
	name     string
	helptext string
}

func (c *creator) Counter() metrics.Counter {
	counter, err := c.meter.NewFloat64Counter(c.name, metric.WithDescription(c.helptext))
	if err != nil {
		log.Warnf("registering counter %s: %s", c.name, err)
		return nil
	}

	return &otelCounter{counter: counter}
}

type otelCounter struct {
	counter metric.Float64Counter
}

func (oc *otelCounter) Inc() {
	oc.Add(1)
}

func (oc *otelCounter) Add(v float64) {
	if oc == nil {
		return
	}
	oc.counter.Add(context.Background(), v)
}

func (c *creator) Gauge() metrics.Gauge {
	panic("not supported")
}

func (c *creator) Histogram(buckets []float64) metrics.Histogram {
	valueRecorder, err := c.meter.NewFloat64ValueRecorder(c.name, metric.WithDescription(c.helptext))
	if err != nil {
		log.Warnf("registering histogram %s: %s", c.name, err)
		return nil
	}
	return &otelHistogram{histogram: valueRecorder}
}

type otelHistogram struct {
	histogram metric.Float64ValueRecorder
}

func (oh *otelHistogram) Observe(v float64) {
	if oh == nil {
		return
	}
	oh.histogram.Record(context.Background(), v)
}

func (c *creator) Summary(opts metrics.SummaryOpts) metrics.Summary {
	panic("not supported")
}
