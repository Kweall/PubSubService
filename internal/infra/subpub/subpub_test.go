package subpub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub_BasicSubscribePublish(t *testing.T) {
	ps := NewSubPub()
	key := "test"
	expectedMsg := "hello"

	var wg sync.WaitGroup
	wg.Add(1)

	_, err := ps.Subscribe(key, func(msg interface{}) {
		defer wg.Done()
		assert.Equal(t, expectedMsg, msg)
	})
	require.NoError(t, err)

	err = ps.Publish(key, expectedMsg)
	require.NoError(t, err)

	wg.Wait() // Ожидаем получение сообщения
}

func TestPubSub_MultipleSubscribers(t *testing.T) {
	ps := NewSubPub()
	key := "test"
	msg := "event"
	const subscribers = 5

	var wg sync.WaitGroup
	wg.Add(subscribers)

	for i := 0; i < subscribers; i++ {
		_, err := ps.Subscribe(key, func(msg interface{}) {
			defer wg.Done()
		})
		require.NoError(t, err)
	}

	err := ps.Publish(key, msg)
	require.NoError(t, err)

	wg.Wait() // Все подписчики должны получить сообщение
}

func TestPubSub_Unsubscribe(t *testing.T) {
	ps := NewSubPub()
	key := "test"

	received := false
	sub, err := ps.Subscribe(key, func(msg interface{}) {
		received = true
	})
	require.NoError(t, err)

	sub.Unsubscribe()
	err = ps.Publish(key, "data")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.False(t, received, "Unsubscribed handler should not be called")
}

func TestPubSub_ConcurrentAccess(t *testing.T) {
	ps := NewSubPub()
	key := "test"
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numOps * 2)

	// Конкурентные подписки
	for i := 0; i < numOps; i++ {
		go func() {
			defer wg.Done()
			_, _ = ps.Subscribe(key, func(msg interface{}) {})
		}()
	}

	// Конкурентные публикации
	for i := 0; i < numOps; i++ {
		go func() {
			defer wg.Done()
			_ = ps.Publish(key, "data")
		}()
	}

	wg.Wait()
}

func TestPubSub_Close(t *testing.T) {
	ps := NewSubPub()
	key := "test"

	// Добавляем активную подписку
	_, err := ps.Subscribe(key, func(msg interface{}) {})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = ps.Close(ctx)
	assert.NoError(t, err)

	// Проверяем что система закрыта
	err = ps.Publish(key, "data")
	assert.Error(t, err)
}

func TestPubSub_FIFOOrder(t *testing.T) {
	ps := NewSubPub()
	key := "test"
	messages := []string{"msg1", "msg2", "msg3"}
	var received []string

	var wg sync.WaitGroup
	wg.Add(len(messages))

	_, err := ps.Subscribe(key, func(msg interface{}) {
		defer wg.Done()
		received = append(received, msg.(string))
	})
	require.NoError(t, err)

	for _, msg := range messages {
		err := ps.Publish(key, msg)
		require.NoError(t, err)
	}

	wg.Wait()
	assert.Equal(t, messages, received, "Messages should be received in FIFO order")
}

func TestPubSub_SlowSubscriber(t *testing.T) {
	ps := NewSubPub()
	key := "test"
	fastReceived := false

	// Медленный подписчик
	_, err := ps.Subscribe(key, func(msg interface{}) {
		time.Sleep(500 * time.Millisecond)
	})
	require.NoError(t, err)

	// Быстрый подписчик
	_, err = ps.Subscribe(key, func(msg interface{}) {
		fastReceived = true
	})
	require.NoError(t, err)

	err = ps.Publish(key, "data")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.True(t, fastReceived, "Fast subscriber should process message immediately")
}

func TestPubSub_PublishBeforeSubscribe(t *testing.T) {
	ps := NewSubPub()
	key := "test"
	msg := "hello"

	err := ps.Publish(key, msg)
	require.NoError(t, err)

	received := make(chan bool, 1)
	_, err = ps.Subscribe(key, func(m interface{}) {
		if m == msg {
			received <- true
		}
	})
	require.NoError(t, err)

	select {
	case <-received:
		t.Fatal("Message should not be received")
	case <-time.After(100 * time.Millisecond):
	}
}
