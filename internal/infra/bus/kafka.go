package bus

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
)

// kafkaMessage Kafka 原始消息
type kafkaMessage struct {
	Key     []byte
	Value   []byte
	Headers map[string]string

	// 用于手动提交的字段
	Topic     string
	Partition int
	Offset    int64
}

type kafkaClient struct {
	bootstrapServers []string
	defaultTopic     string
	defaultGroupID   string
	writers          map[string]*kafkago.Writer // topic -> writer
	readers          map[string]*kafkago.Reader // topic -> reader
	reader           *kafkago.Reader            // 默认 reader（兼容单 topic）
	mu               sync.RWMutex
}

// ConsumerMode 消费者模式
type ConsumerMode int

const (
	// ConsumerModeShared 共享模式：多实例共享同一 GroupID
	ConsumerModeShared ConsumerMode = iota
	// ConsumerModeBroadcast 广播模式：每个实例唯一 GroupID，实例间全量消费
	ConsumerModeBroadcast
)

// ConsumerOption 消费者配置选项
type ConsumerOption func(*consumerOptions)

type consumerOptions struct {
	mode    ConsumerMode
	groupID string
}

// WithConsumerMode 设置消费者模式
func WithConsumerMode(mode ConsumerMode) ConsumerOption {
	return func(o *consumerOptions) {
		o.mode = mode
	}
}

// WithGroupID 设置消费者组
func WithGroupID(groupID string) ConsumerOption {
	return func(o *consumerOptions) {
		o.groupID = groupID
	}
}

// newClient 从配置创建 Kafka 客户端
func newClient(cfg *config) *kafkaClient {
	defaultTopic, err := cfg.TopicName(TopicDefault)
	if err != nil {
		panic(err)
	}

	return &kafkaClient{
		bootstrapServers: append([]string(nil), cfg.BootstrapServers...),
		defaultTopic:     defaultTopic,
		defaultGroupID:   cfg.GroupID,
		writers:          make(map[string]*kafkago.Writer),
		readers:          make(map[string]*kafkago.Reader),
	}
}

// InitReader 初始化默认 topic 的 reader
func (c *kafkaClient) InitReader(opts ...ConsumerOption) {
	c.InitReaderForTopics([]string{c.defaultTopic}, opts...)
}

// InitReaderForTopics 初始化指定 topics 的 reader
func (c *kafkaClient) InitReaderForTopics(topics []string, opts ...ConsumerOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭旧 readers
	for _, r := range c.readers {
		_ = r.Close()
	}
	c.readers = make(map[string]*kafkago.Reader)

	if c.reader != nil {
		_ = c.reader.Close()
		c.reader = nil
	}

	options := &consumerOptions{
		mode:    ConsumerModeShared,
		groupID: c.defaultGroupID,
	}
	for _, opt := range opts {
		opt(options)
	}

	groupID := options.groupID
	if options.mode == ConsumerModeBroadcast {
		hostname, _ := os.Hostname()
		groupID = fmt.Sprintf("%s-%s-%s", c.defaultGroupID, hostname, uuid.New().String()[:8])
	}

	for _, topic := range topics {
		reader := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        c.bootstrapServers,
			GroupID:        groupID,
			Topic:          topic,
			CommitInterval: time.Second,
			MinBytes:       1,
			MaxBytes:       10e6,
		})
		c.readers[topic] = reader
	}

	if len(topics) > 0 {
		c.reader = c.readers[topics[0]]
	}
}

// Write 写入默认 topic
func (c *kafkaClient) Write(ctx context.Context, msg kafkaMessage) error {
	return c.WriteTo(ctx, c.defaultTopic, msg)
}

// WriteTo 写入指定 topic
func (c *kafkaClient) WriteTo(ctx context.Context, topic string, msg kafkaMessage) error {
	writer := c.getOrCreateWriter(topic)

	kafkaMsg := kafkago.Message{
		Key:   msg.Key,
		Value: msg.Value,
	}

	if msg.Headers != nil {
		kafkaMsg.Headers = make([]kafkago.Header, 0, len(msg.Headers))
		for key, value := range msg.Headers {
			kafkaMsg.Headers = append(kafkaMsg.Headers, kafkago.Header{
				Key:   key,
				Value: []byte(value),
			})
		}
	}

	return writer.WriteMessages(ctx, kafkaMsg)
}

func (c *kafkaClient) getOrCreateWriter(topic string) *kafkago.Writer {
	c.mu.RLock()
	if writer, ok := c.writers[topic]; ok {
		c.mu.RUnlock()
		return writer
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if writer, ok := c.writers[topic]; ok {
		return writer
	}

	writer := &kafkago.Writer{
		Addr:                   kafkago.TCP(c.bootstrapServers...),
		Topic:                  topic,
		Balancer:               &kafkago.Hash{},
		BatchTimeout:           10 * time.Millisecond,
		RequiredAcks:           kafkago.RequireAll,
		BatchBytes:             10 * 1024 * 1024,
		AllowAutoTopicCreation: true,
	}
	c.writers[topic] = writer

	return writer
}

// Read 从默认 reader 读取（ReadMessage，自动提交）
func (c *kafkaClient) Read(ctx context.Context) (kafkaMessage, error) {
	if c.reader == nil {
		return kafkaMessage{}, fmt.Errorf("kafka reader is not initialized")
	}

	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return kafkaMessage{}, err
	}

	return fromKafkaMessage(msg), nil
}

// ReadFrom 从指定 topic reader 读取（ReadMessage，自动提交）
func (c *kafkaClient) ReadFrom(ctx context.Context, topic string) (kafkaMessage, error) {
	c.mu.RLock()
	reader, ok := c.readers[topic]
	c.mu.RUnlock()

	if !ok || reader == nil {
		return kafkaMessage{}, fmt.Errorf("kafka reader for topic %s is not initialized", topic)
	}

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		return kafkaMessage{}, err
	}

	return fromKafkaMessage(msg), nil
}

// FetchFrom 从指定 topic reader 拉取消息（FetchMessage，需手动提交）
func (c *kafkaClient) FetchFrom(ctx context.Context, topic string) (kafkaMessage, error) {
	c.mu.RLock()
	reader, ok := c.readers[topic]
	c.mu.RUnlock()

	if !ok || reader == nil {
		return kafkaMessage{}, fmt.Errorf("kafka reader for topic %s is not initialized", topic)
	}

	msg, err := reader.FetchMessage(ctx)
	if err != nil {
		return kafkaMessage{}, err
	}

	return fromKafkaMessage(msg), nil
}

// Commit 提交已拉取消息
func (c *kafkaClient) Commit(ctx context.Context, msg kafkaMessage) error {
	if msg.Topic != "" {
		c.mu.RLock()
		reader, ok := c.readers[msg.Topic]
		c.mu.RUnlock()
		if !ok || reader == nil {
			return fmt.Errorf("kafka reader for topic %s is not initialized", msg.Topic)
		}
		return reader.CommitMessages(ctx, kafkago.Message{
			Topic:     msg.Topic,
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Key:       msg.Key,
			Value:     msg.Value,
		})
	}

	if c.reader == nil {
		return fmt.Errorf("kafka reader is not initialized")
	}

	return c.reader.CommitMessages(ctx, kafkago.Message{
		Key:   msg.Key,
		Value: msg.Value,
	})
}

// Consume 创建独立 reader（shared 模式），阻塞运行消费循环
func (c *kafkaClient) Consume(ctx context.Context, topic, groupID string, handler func(context.Context, []byte) error) error {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        c.bootstrapServers,
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: time.Second,
		MinBytes:       1,
		MaxBytes:       10e6,
	})
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				time.Sleep(time.Second)
				continue
			}
			if err := handler(ctx, msg.Value); err != nil {
				return err
			}
			if err := reader.CommitMessages(ctx, msg); err != nil {
				return err
			}
		}
	}
}

// ConsumeBroadcast 创建独立 reader（broadcast 模式），阻塞运行消费循环
func (c *kafkaClient) ConsumeBroadcast(ctx context.Context, topic string, handler func(context.Context, []byte) error) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	groupID := fmt.Sprintf("%s-broadcast-%s-%s", c.defaultGroupID, hostname, uuid.New().String()[:8])
	return c.Consume(ctx, topic, groupID, handler)
}

// Close 关闭客户端连接
func (c *kafkaClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error

	for _, writer := range c.writers {
		if closeErr := writer.Close(); closeErr != nil {
			err = closeErr
		}
	}
	c.writers = nil

	for _, reader := range c.readers {
		if closeErr := reader.Close(); closeErr != nil {
			err = closeErr
		}
	}
	c.readers = nil

	if c.reader != nil {
		if closeErr := c.reader.Close(); closeErr != nil {
			err = closeErr
		}
		c.reader = nil
	}

	return err
}

func (c *kafkaClient) CreateTopics(ctx context.Context, topics []topic) error {
	if len(topics) == 0 {
		return nil
	}

	brokerConn, err := c.dialBroker(ctx)
	if err != nil {
		return err
	}
	defer brokerConn.Close()

	controller, err := brokerConn.Controller()
	if err != nil {
		return fmt.Errorf("kafka get handler failed: %w", err)
	}

	controllerAddress := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := kafkago.DialContext(ctx, "tcp", controllerAddress)
	if err != nil {
		return fmt.Errorf("kafka dial handler failed: %w", err)
	}
	defer controllerConn.Close()

	configs := make([]kafkago.TopicConfig, 0, len(topics))
	for _, item := range topics {
		if item.Name == "" {
			continue
		}
		configs = append(configs, kafkago.TopicConfig{
			Topic:             item.Name,
			NumPartitions:     item.Partitions,
			ReplicationFactor: item.ReplicationFactor,
		})
	}

	if len(configs) == 0 {
		return nil
	}

	if err := controllerConn.CreateTopics(configs...); err != nil {
		return fmt.Errorf("kafka create topics failed: %w", err)
	}

	return nil
}

func (c *kafkaClient) dialBroker(ctx context.Context) (*kafkago.Conn, error) {
	var lastErr error
	for _, server := range c.bootstrapServers {
		conn, err := kafkago.DialContext(ctx, "tcp", server)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no kafka bootstrap server configured")
	}

	return nil, fmt.Errorf("kafka dial broker failed: %w", lastErr)
}

func fromKafkaMessage(msg kafkago.Message) kafkaMessage {
	headers := make(map[string]string, len(msg.Headers))
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	return kafkaMessage{
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   headers,
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}
}
