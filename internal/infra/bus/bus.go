package bus

import (
	"context"
)

type Bus struct {
	pub  *publisher
	cons *consumer
	cfg  *config
}

func NewBus() *Bus {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	client := newClient(cfg)
	pub := newPublisher(client)
	cons := newConsumer(client, pub)
	cons.use(recovery(), logging())
	return &Bus{
		pub:  pub,
		cons: cons,
		cfg:  cfg,
	}
}

func (b *Bus) Publish(ctx context.Context, event Event) error {
	topic, err := b.cfg.TopicName(event.Topic())
	if err != nil {
		return ErrEventTopic.WithCause(err)
	}

	return b.pub.publish(ctx, event, topic)
}

func (b *Bus) Listen(pattern string, handler Handler) {
	b.cons.listen(pattern, handler)
}

func (b *Bus) Register(listeners ...Listener) {
	for _, listener := range listeners {
		for pattern, handler := range listener.Handlers() {
			b.Listen(pattern, handler)
		}
	}
}

func (b *Bus) Use(middleware ...Middleware) {
	b.cons.use(middleware...)
}

func (b *Bus) Init(consumer string, opts ...ConsumerOption) error {
	rawTopics, err := b.cfg.ConsumerTopicNames(consumer)
	if err != nil {
		return err
	}

	b.cons.init(rawTopics, opts...)
	return nil
}

func (b *Bus) EnableDLQ(resolver func(topic string) string) {
	b.cons.mu.Lock()
	b.cons.dlqTopic = resolver
	b.cons.mu.Unlock()
}

func (b *Bus) Run(ctx context.Context) error {
	return b.cons.run(ctx)
}

func (b *Bus) CreateAllTopics(ctx context.Context, includeDLQ bool) error {
	topics := b.cfg.TopicsForProvision()
	if includeDLQ {
		dlqTopics := make([]topic, 0, len(topics))
		for _, item := range topics {
			dlqTopics = append(dlqTopics, topic{
				Name:              item.Name + ".dlq",
				Partitions:        item.Partitions,
				ReplicationFactor: item.ReplicationFactor,
			})
		}
		topics = append(topics, dlqTopics...)
	}

	return b.pub.kafka.CreateTopics(ctx, topics)
}
