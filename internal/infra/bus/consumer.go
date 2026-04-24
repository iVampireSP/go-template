package bus

import (
	"context"
	"sync"
	"time"

	"github.com/iVampireSP/go-template/pkg/cerr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var (
	ErrDecodeMessage = cerr.Internal("failed to decode event message").WithCode("BUS_DECODE_EVENT")
	ErrProcessEvent  = cerr.Internal("failed to process event").WithCode("BUS_PROCESS_EVENT")
	ErrCommitEvent   = cerr.Internal("failed to commit event message").WithCode("BUS_COMMIT_EVENT")
	ErrTopicRequired = cerr.BadRequest("consumer topics are required").WithCode("BUS_CONSUMER_TOPICS_REQUIRED")
)

type consumer struct {
	kafka      *kafkaClient
	publisher  *publisher
	middleware []Middleware
	listeners  map[string][]Handler
	topics     []string
	initOpts   []ConsumerOption
	dlqTopic   func(topic string) string
	mu         sync.RWMutex
	inited     bool
}

func newConsumer(kafka *kafkaClient, pub *publisher) *consumer {
	return &consumer{
		kafka:     kafka,
		publisher: pub,
		listeners: map[string][]Handler{},
	}
}

func (c *consumer) listen(pattern string, handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.listeners[pattern] = append(c.listeners[pattern], handler)
}

func (c *consumer) use(middleware ...Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.middleware = append(c.middleware, middleware...)
}

func (c *consumer) init(topics []string, opts ...ConsumerOption) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.topics = topics
	c.initOpts = opts
}

func (c *consumer) run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.mu.Lock()
	topics := append([]string(nil), c.topics...)
	if len(topics) == 0 {
		c.mu.Unlock()
		return ErrTopicRequired
	}
	if !c.inited {
		c.kafka.InitReaderForTopics(topics, c.initOpts...)
		c.inited = true
	}
	c.mu.Unlock()

	go c.kafka.StartStatsCollector(runCtx)

	msgCh := make(chan kafkaMessage, 100)
	var wg sync.WaitGroup
	for _, topic := range topics {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			for {
				select {
				case <-runCtx.Done():
					return
				default:
					msg, err := c.kafka.FetchFrom(runCtx, t)
					if err != nil {
						if runCtx.Err() != nil {
							return
						}
						select {
						case <-runCtx.Done():
							return
						case <-time.After(time.Second):
						}
						continue
					}
					select {
					case msgCh <- msg:
					case <-runCtx.Done():
						return
					}
				}
			}
		}(topic)
	}

	go func() {
		wg.Wait()
		close(msgCh)
	}()

	for {
		select {
		case <-runCtx.Done():
			return runCtx.Err()
		case msg, ok := <-msgCh:
			if !ok {
				return nil
			}
			shouldCommit, err := c.handleMessage(runCtx, msg)
			if err != nil {
				if !shouldCommit {
					cancel()
					return err
				}
				if commitErr := c.kafka.Commit(runCtx, msg); commitErr != nil {
					cancel()
					return ErrCommitEvent.WithCause(commitErr)
				}
				continue
			}
			if !shouldCommit {
				continue
			}
			if commitErr := c.kafka.Commit(runCtx, msg); commitErr != nil {
				cancel()
				return ErrCommitEvent.WithCause(commitErr)
			}
		}
	}
}

func (c *consumer) handleMessage(ctx context.Context, msg kafkaMessage) (bool, error) {
	env, err := decodeEnvelope(msg.Value)
	if err != nil {
		return false, ErrDecodeMessage.WithCause(err)
	}

	if len(env.Metadata) > 0 {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(env.Metadata))
	}
	ctx = ContextWithEnvelope(ctx, env)

	c.mu.RLock()
	var handlers []Handler
	for pattern, hs := range c.listeners {
		if matchPattern(pattern, env.Name) {
			handlers = append(handlers, hs...)
		}
	}
	dlq := c.dlqTopic
	middleware := append([]Middleware(nil), c.middleware...)
	c.mu.RUnlock()

	if len(handlers) == 0 {
		return true, nil
	}

	var firstErr error
	for _, h := range handlers {
		handler := h
		for idx := len(middleware) - 1; idx >= 0; idx-- {
			handler = middleware[idx](handler)
		}
		if err := handler(ctx, env.Payload); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if firstErr != nil {
		if dlq == nil {
			return false, ErrProcessEvent.WithCause(firstErr)
		}
		dlqTopic := dlq(msg.Topic)
		if dlqTopic == "" {
			return false, ErrProcessEvent.WithCause(firstErr)
		}
		if dlqErr := c.publisher.publishDLQ(ctx, dlqTopic, env.Key, newDLQEnvelope(env, firstErr)); dlqErr != nil {
			return false, ErrProcessEvent.WithMessage("failed to publish event to dlq").WithCause(dlqErr)
		}
		return true, ErrProcessEvent.WithCause(firstErr)
	}
	return true, nil
}
