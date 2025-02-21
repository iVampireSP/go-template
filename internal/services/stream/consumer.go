package stream

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

//func (s *Service) Listen(topic string, handler HandlerFunc) error {
//	conn, err := s.dial(topic)
//	if err != nil {
//		return err
//	}
//
//	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
//
//	b := make([]byte, 10e3) // 10KB max per message
//	for {
//		n, err := batch.Read(b)
//		if err != nil {
//			return err
//		}
//		handler(b[:n])
//	}
//}

func (s *Service) Consumer(topic string, groupId string) *kafka.Reader {
	var r = kafka.ReaderConfig{
		Brokers:        s.config.Kafka.BootstrapServers,
		GroupID:        groupId,
		GroupTopics:    nil,
		Topic:          topic,
		CommitInterval: time.Second,
		//StartOffset:            kafka., // 仅对新创建的消费者组生效，从头开始消费，工作中可能更常用从最新的开始消费kafka.LastOffset
		Logger:                nil,
		ErrorLogger:           nil,
		IsolationLevel:        0,
		MaxAttempts:           0,
		OffsetOutOfRangeError: false,
		MaxBytes:              10e6, // 10MB
	}

	if s.config.Kafka.Username != "" && s.config.Kafka.Password != "" {
		r.Dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: s.auth(),
		}
	}

	return kafka.NewReader(r)
}

type HandlerFunc func([]byte)

// ReadMessage 消费消息
func (s *Service) ReadMessage(ctx context.Context, topic string, groupId string) {
	for {
		if msg, err := s.Consumer(topic, groupId).ReadMessage(ctx); err != nil {
			fmt.Println(fmt.Sprintf("读kafka失败，err:%v", err))
			continue
		} else {
			fmt.Println(fmt.Sprintf("topic=%s,partition=%d,offset=%d,key=%s,value=%s", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value))
		}
	}
}
