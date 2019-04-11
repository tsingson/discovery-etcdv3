package tracing

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// TracerProvider enum
type Provider int

const (
	// Jaeger provider
	Jaeger Provider = iota
	// Zipkin provider
	Zipkin
)

// TracerConfig for tracer
type TracerConfig struct {
	Provider    Provider
	ServiceName string
	Host        string
	Port        string
}

// NewTracer via TracerConfig, returns opentracing.Tracer
func NewTracer(cfg TracerConfig) opentracing.Tracer {
	var tracer opentracing.Tracer
	switch cfg.Provider {
	case Zipkin:
		logrus.Error("No implements yet.")
		// fmt.Sprintf("http://%s:%s/api/v1/spans",cfg.Host, cfg.Port)
		break
	case Jaeger:
		tracer = newJaegerTracer(cfg)
		break
	default:
		logrus.Errorf("unsported provider %s, use opentracing.GlobalTracer()", cfg.Provider)
		tracer = opentracing.GlobalTracer()
	}
	return tracer
}

func newJaegerTracer(tc TracerConfig) opentracing.Tracer {
	cfg := jaegercfg.Configuration{
		ServiceName: tc.ServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", tc.Host, tc.Port),
		},
	}
	logrus.Infof("Using Jaeger HTTP tracer: %s", fmt.Sprintf("%s:%s", tc.Host, tc.Port))

	tracer, _, err := cfg.NewTracer()
	if err != nil {
		logrus.Fatalf("unable to create Jaeger tracer: %+v", err)
	}
	if tracer == nil {
		tracer = opentracing.GlobalTracer()
	}

	return tracer
}
