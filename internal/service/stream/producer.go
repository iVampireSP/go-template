package stream

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go-template/internal/schema"
	"time"
)

//var connections = map[string]*kafka.Conn{}

//func (s *Service) dial(topic string) (*kafka.Conn, error) {
//
//	// 如果topic 存在于 connections 则直接返回
//	if conn, ok := connections[topic]; ok {
//		return conn, nil
//	}
//
//	ctx := context.Background()
//
//	conn, err := kafka.DialLeader(ctx, "tcp", s.config.Kafka.BootstrapServers[0], s.topic(topic), 0)
//	if err != nil {
//		return nil, err
//	}
//	//err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
//	//if err != nil {
//	//	return conn, err
//	//}
//	//
//	//// set read deadline
//	//err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
//	//if err != nil {
//	//	return conn, err
//	//}
//
//	connections[topic] = conn
//
//	return conn, nil
//}
//
//func (s *Service) Publish(topic string, message ...[]byte) error {
//	conn, err := s.dial(topic)
//	if err != nil {
//		return err
//	}
//
//	msg := make([]kafka.Message, len(message))
//	for i, v := range message {
//		msg[i] = kafka.Message{Value: v}
//	}
//
//	_, err = conn.WriteMessages(msg...)
//
//	return err
//}

func (s *Service) Producer(topic string) *kafka.Writer {
	var w = &kafka.Writer{
		Addr:            kafka.TCP(s.config.Kafka.BootstrapServers...),
		Topic:           topic,
		Balancer:        &kafka.Hash{}, // 用于对key进行hash，决定消息发送到哪个分区
		MaxAttempts:     0,
		WriteBackoffMin: 0,
		WriteBackoffMax: 0,
		BatchSize:       0,
		BatchBytes:      0,
		BatchTimeout:    0,
		ReadTimeout:     0,
		//WriteTimeout:           time.Second,       // kafka有时候可能负载很高，写不进去，那么超时后可以放弃写入，用于可以丢消息的场景
		RequiredAcks:           kafka.RequireAll, // 不需要任何节点确认就返回
		Async:                  true,
		Completion:             nil,
		Compression:            0,
		Logger:                 nil,
		ErrorLogger:            nil,
		Transport:              nil,
		AllowAutoTopicCreation: false, // 第一次发消息的时候，如果topic不存在，就自动创建topic，工作中禁止使用
	}

	if s.config.Kafka.Username != "" && s.config.Kafka.Password != "" {
		w.Transport = &kafka.Transport{
			SASL: s.auth(),
		}
	}

	return w
}

func (s *Service) SendMessage(ctx context.Context, topic string, data []byte) error {
	msg := kafka.Message{
		Partition:     0,
		Offset:        0,
		HighWaterMark: 0,
		//Key:           key,
		Value: data,
		Time:  time.Time{},
	}

	err := s.Producer(topic).WriteMessages(ctx, msg)
	return err
}

func (s *Service) SendEvent(ctx context.Context, topic string, data schema.EventMessage) error {
	j, err := data.JSON()
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Partition:     0,
		Offset:        0,
		HighWaterMark: 0,
		Value:         j,
		Time:          time.Time{},
	}

	err = s.Producer(topic).WriteMessages(ctx, msg)
	return err
}
