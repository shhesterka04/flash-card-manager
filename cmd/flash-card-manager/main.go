package main

import (
	"context"
	"fmt"
	pb "flash-card-manager/internal/app/grpc"
	"flash-card-manager/internal/app/handlers"
	"flash-card-manager/internal/app/handlers/utils"
	"flash-card-manager/internal/infrastructure/kafka"
	"flash-card-manager/pkg/db"
	"flash-card-manager/pkg/logger"
	repository "flash-card-manager/pkg/repository/init"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	otgrpc "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

const (
	port = ":9000"
)


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Init()

	tracer, closer := utils.InitJaeger("flash-card-manager-service")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	database, err := db.NewDB(ctx, db.GenerateDsn())
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize database: %v", err)
	}
	defer database.GetPool(ctx).Close()

	cardRepo, deckRepo := repository.InitRepositories(database)

	producer, _, err := kafka.InitializeKafka()
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize Kafka: %v", err)
	}
	defer producer.Close()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			otgrpc.UnaryServerInterceptor(otgrpc.WithTracer(opentracing.GlobalTracer())),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			otgrpc.StreamServerInterceptor(otgrpc.WithTracer(opentracing.GlobalTracer())),
		)),
	)

	deckHandler := handlers.NewDeckServiceServer(deckRepo, kafka.NewKafkaEventSender(producer))
	cardHandler := handlers.NewCardServiceServer(cardRepo, kafka.NewKafkaEventSender(producer))

	pb.RegisterDeckServiceServer(grpcServer, deckHandler)
	pb.RegisterCardServiceServer(grpcServer, cardHandler)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		logger.Errorf(ctx, "Failed to listen on port %s: %v", port, err)
	}

	fmt.Printf("Server listening at %v", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		logger.Errorf(ctx, "Failed to serve gRPC server over port %s: %v", port, err)
	}
}
