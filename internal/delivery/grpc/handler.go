package grpc

import (
	"context"
	"log"

	"pubsub_service/internal/app"
	gen "pubsub_service/internal/delivery/grpc/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PubSubHandler struct {
	gen.UnimplementedPubSubServer
	svc *app.PubSubService
}

func NewPubSubHandler(svc *app.PubSubService) *PubSubHandler {
	return &PubSubHandler{svc: svc}
}

func (h *PubSubHandler) Subscribe(req *gen.SubscribeRequest, stream gen.PubSub_SubscribeServer) error {
	if req.GetKey() == "" {
		return status.Error(codes.InvalidArgument, "key is required")
	}

	ctx := stream.Context()
	return h.svc.Subscribe(ctx, req.GetKey(), func(data string) {
		_ = stream.Send(&gen.Event{Data: data}) // потоковое отправление
	})
}

func (h *PubSubHandler) Publish(ctx context.Context, req *gen.PublishRequest) (*emptypb.Empty, error) {
	if req.GetKey() == "" || req.GetData() == "" {
		return nil, status.Error(codes.InvalidArgument, "key and data required")
	}

	err := h.svc.Publish(ctx, req.GetKey(), req.GetData())
	if err != nil {
		log.Println("Publish error:", err)
		return nil, status.Error(codes.Internal, "failed to publish")
	}

	return &emptypb.Empty{}, nil
}
