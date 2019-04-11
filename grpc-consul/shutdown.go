package gf

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// Shutdown service server instance
func (service *Service) Shutdown() {
	logrus.Infof("Gracefully shutting down gRPC and Prometheus")
	if service.Registry != nil {
		service.Registry.DeRegister(service.Name)
	}
	if service.grpcServer != nil {
		service.grpcServer.GracefulStop()
	}
	if service.prometheusServer != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancel()

		if err := service.prometheusServer.Shutdown(ctx); err != nil {
			logrus.Infof("Timeout during shutdown of metrics server. Error: %v", err)
		}
	}

}
