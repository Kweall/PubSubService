package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pubsub_service/internal/app"
	delivery "pubsub_service/internal/delivery/grpc"
	"pubsub_service/internal/infra/subpub"
	"pubsub_service/pkg/config"

	gen "pubsub_service/internal/delivery/grpc/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	// Инициализация pubsub-системы
	subpubSystem := subpub.NewSubPub()
	service := app.NewPubSubService(subpubSystem)
	handler := delivery.NewPubSubHandler(service)

	// gRPC сервер
	grpcServer := grpc.NewServer()
	gen.RegisterPubSubServer(grpcServer, handler)

	reflection.Register(grpcServer)
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Graceful shutdown
	go func() {
		log.Printf("gRPC server listening on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Завершение по Ctrl+C или SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := subpubSystem.Close(ctx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
}
