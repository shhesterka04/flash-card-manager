package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	pb "flash-card-manager/internal/app/grpc"
	"flash-card-manager/internal/app/handlers/utils"
	"flash-card-manager/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	logger.Init()

	addr := flag.String("addr", "localhost:9000", "the address to connect to the gRPC server")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: go run cmd/client/main.go -addr=localhost:9000 [command] [data]")
		os.Exit(1)
	}

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf(ctx, "Could not connect to %s: %v", *addr, err)
		os.Exit(1)
	}
	defer conn.Close()

	deckClient := pb.NewDeckServiceClient(conn)
	cardClient := pb.NewCardServiceClient(conn)

	if err := utils.HandleCommand(ctx, deckClient, cardClient, flag.Arg(0), flag.Args()[1:]); err != nil {
		logger.Errorf(ctx, "Error handling command: %v", err)
		os.Exit(1)
	}
}
