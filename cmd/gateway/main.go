package main

import (
	"context"
	pb "flash-card-manager/internal/app/grpc"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}

	mux := runtime.NewServeMux()
	if err := pb.RegisterCardServiceHandler(ctx, mux, conn); err != nil {
		log.Fatalf("failed to register card service handler: %v", err)
	}
	if err := pb.RegisterDeckServiceHandler(ctx, mux, conn); err != nil {
		log.Fatalf("failed to register deck service handler: %v", err)
	}

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
