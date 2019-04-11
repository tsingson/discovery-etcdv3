package gf

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
)

// Run service
func (service *Service) Run() {
	// check registry
	if service.Registry != nil {
		err := service.Registry.Register(service.ID, service.Name, service.ServiceAdrr.Port, "rpc-endpoint")
		if err != nil {
			logrus.Errorf("Failed to register service, Error: %v", err)
		}
	}

	// Promethues
	startPrometheusServer(service)

	// grpc
	go func() {
		if err := createGrpcServer(service); err != nil {
			logrus.Fatalf("gRPC serve failed. Error: %v", err)
		}
	}()

	gracefulShutdown(service)
}

func createGrpcServer(service *Service) error {

	listener, err := net.Listen("tcp", service.ServiceAdrr.String())
	if err != nil {
		return err
	}
	logrus.Infof("Serving gRPC on %s", service.ServiceAdrr.String())

	service.AddUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor)
	service.AddStreamInterceptor(grpc_prometheus.StreamServerInterceptor)

	if service.Tracer == nil {
		service.Tracer = opentracing.GlobalTracer()
	}
	service.AddUnaryInterceptor(otgrpc.OpenTracingServerInterceptor(service.Tracer))
	service.AddStreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(service.Tracer))

	service.AddStreamInterceptor(grpc_recovery.StreamServerInterceptor())
	service.AddUnaryInterceptor(grpc_recovery.UnaryServerInterceptor())

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(service.UnaryInterceptors...)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(service.StreamInterceptors...)),
	)

	service.GRPCServer(server)

	grpc_prometheus.EnableHandlingTimeHistogram(
		func(opt *prometheus.HistogramOpts) {
			opt.Buckets = prometheus.ExponentialBuckets(0.005, 1.4, 20)
		},
	)

	grpc_prometheus.Register(server)

	// for shutdown
	service.grpcServer = server

	return server.Serve(listener)
}

func startPrometheusServer(service *Service) {
	s := &http.Server{Addr: service.PrometheusAdrr.String()}
	// for shutdown
	service.prometheusServer = s

	http.Handle("/metrics", promhttp.Handler())
	logrus.Infof("Prometheus metrics at http://%s/metrics", service.PrometheusAdrr.String())
	go func() {
		if err := s.ListenAndServe(); err != nil {
			logrus.Errorf("Prometheus http server: ListenAndServe() error: %s", err)
		}
	}()
}

func gracefulShutdown(service *Service) {
	// error chan
	ec := make(chan error, 10)
	// signal chan
	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-ec:
			if err != nil {
				logrus.Fatalf("Error during application: %v", err)
			}
		case s := <-sc:
			logrus.Infof("Captured %v. Exiting...", s)
			service.Shutdown()
			os.Exit(0)
		}
	}
}
