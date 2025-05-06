package main

import (
	"context"
	"log"

	pb "pubsub_service/internal/delivery/grpc/gen"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPubSubClient(conn)

	// Подписка
	stream, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{Key: "test"})
	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	go func() {
		for {
			event, err := stream.Recv()
			if err != nil {
				log.Printf("Stream error: %v", err)
				return
			}
			log.Printf("Received: %s", event.Data)
		}
	}()

	// Публикация
	_, err = client.Publish(context.Background(), &pb.PublishRequest{
		Key:  "test",
		Data: "Hello from client",
	})
	if err != nil {
		log.Fatalf("Publish failed: %v", err)
	}

	// Ожидание событий
	select {}
}
