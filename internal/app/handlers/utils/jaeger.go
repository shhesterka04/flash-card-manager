package utils

import (
	"io"
	"time"

	"flash-card-manager/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func InitJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, err := cfg.New(
		service,
		config.Logger(jaeger.StdLogger),
	)

	if err != nil {
		logger.GetLogger().Sugar().Errorf("ERROR: cannot init Jaeger: %v\n", err)
	}
	return tracer, closer
}
