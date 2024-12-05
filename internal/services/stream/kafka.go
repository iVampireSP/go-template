package stream

import (
	"github.com/segmentio/kafka-go/sasl/plain"
)

func (s *Service) auth() plain.Mechanism {
	mechanism := plain.Mechanism{
		Username: s.config.Kafka.Username,
		Password: s.config.Kafka.Password,
	}
	return mechanism
}
