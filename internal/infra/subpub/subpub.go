package subpub

import (
    "context"
    "errors"
    "sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
    Unsubscribe()
}

type SubPub interface {
    Subscribe(subject string, cb MessageHandler) (Subscription, error)
    Publish(subject string, msg interface{}) error
    Close(ctx context.Context) error
}

type subscriber struct {
    cb     MessageHandler
    ch     chan interface{}
    closed chan struct{}
}

type subscription struct {
    unsub func()
}

func (s *subscription) Unsubscribe() {
    s.unsub()
}

type subpub struct {
    mu          sync.RWMutex
    subscribers map[string]map[*subscriber]struct{}
    closed      chan struct{}
    wg          sync.WaitGroup
}

func NewSubPub() SubPub {
    return &subpub{
        subscribers: make(map[string]map[*subscriber]struct{}),
        closed:      make(chan struct{}),
    }
}

func (s *subpub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
    select {
    case <-s.closed:
        return nil, errors.New("subpub closed")
    default:
    }

    sub := &subscriber{
        cb:     cb,
        ch:     make(chan interface{}, 128),
        closed: make(chan struct{}),
    }

    s.mu.Lock()
    if _, ok := s.subscribers[subject]; !ok {
        s.subscribers[subject] = make(map[*subscriber]struct{})
    }
    s.subscribers[subject][sub] = struct{}{}
    s.mu.Unlock()

    s.wg.Add(1)
    go s.run(sub)

    return &subscription{
        unsub: func() {
            s.mu.Lock()
            if subs, ok := s.subscribers[subject]; ok {
                delete(subs, sub)
                if len(subs) == 0 {
                    delete(s.subscribers, subject)
                }
            }
            s.mu.Unlock()
            close(sub.closed)
        },
    }, nil
}

func (s *subpub) run(sub *subscriber) {
    defer s.wg.Done()
    for {
        select {
        case msg, ok := <-sub.ch:
            if !ok {
                return
            }
            sub.cb(msg)
        case <-sub.closed:
            return
        }
    }
}

func (s *subpub) Publish(subject string, msg interface{}) error {
    select {
    case <-s.closed:
        return errors.New("subpub closed")
    default:
    }

    s.mu.RLock()
    defer s.mu.RUnlock()

    subs, ok := s.subscribers[subject]
    if !ok {
        return nil
    }
    for sub := range subs {
        select {
        case sub.ch <- msg:
        default:
            go func(sub *subscriber) {
                sub.ch <- msg
            }(sub)
        }
    }
    return nil
}

func (s *subpub) Close(ctx context.Context) error {
    select {
    case <-s.closed:
        return nil
    default:
        close(s.closed)
    }

    done := make(chan struct{})

    go func() {
        s.mu.Lock()
        for _, subs := range s.subscribers {
            for sub := range subs {
                close(sub.closed)
                close(sub.ch)
            }
        }
        s.subscribers = nil
        s.mu.Unlock()
        s.wg.Wait()
        close(done)
    }()

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-done:
        return nil
    }
}