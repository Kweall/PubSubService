package app

import (
	"context"
	"pubsub_service/internal/infra/subpub"
)

type PubSubService struct {
	bus subpub.SubPub
}

func NewPubSubService(bus subpub.SubPub) *PubSubService {
	return &PubSubService{bus: bus}
}

func (s *PubSubService) Subscribe(ctx context.Context, key string, send func(data string)) error {
	_, err := s.bus.Subscribe(key, func(msg interface{}) {
		str, ok := msg.(string)
		if !ok {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			send(str)
		}
	})
	if err != nil {
		return err
	}

	<-ctx.Done() // держим подписку живой до отмены
	return nil
}

func (s *PubSubService) Publish(ctx context.Context, key string, data string) error {
	return s.bus.Publish(key, data)
}
