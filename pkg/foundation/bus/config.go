package bus

import (
	"fmt"
	"sort"
	"strings"
)

// Config 事件总线配置，由外部传入。
type Config struct {
	BootstrapServers []string                    `yaml:"bootstrap_servers"`
	GroupID          string                      `yaml:"group_id"`
	Topics           map[string]TopicConfig      `yaml:"topics"`
	Consumers        map[string]ConsumerConfig   `yaml:"consumers"`

	topics map[TopicID]topic
}

// TopicConfig 单个 topic 配置。
type TopicConfig struct {
	Name              string `yaml:"name"`
	Partitions        int    `yaml:"partitions"`
	ReplicationFactor int    `yaml:"replication_factor"`
}

// ConsumerConfig 消费者组配置。
type ConsumerConfig struct {
	Topics []TopicID `yaml:"topics"`
}

type topic struct {
	Name              string
	Partitions        int
	ReplicationFactor int
}

// Validate 校验配置完整性。
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("bus config is nil")
	}

	servers := make([]string, 0, len(c.BootstrapServers))
	for _, server := range c.BootstrapServers {
		server = strings.TrimSpace(server)
		if server == "" {
			continue
		}
		servers = append(servers, server)
	}
	if len(servers) == 0 {
		return fmt.Errorf("bus.bootstrap_servers is required")
	}
	c.BootstrapServers = servers

	c.GroupID = strings.TrimSpace(c.GroupID)
	if c.GroupID == "" {
		return fmt.Errorf("bus.group_id is required")
	}

	if len(c.Topics) == 0 {
		return fmt.Errorf("bus.topics is required")
	}

	topics := make(map[TopicID]topic, len(c.Topics))
	for key, value := range c.Topics {
		topicID := TopicID(strings.TrimSpace(key))
		if topicID == "" {
			return fmt.Errorf("bus.topics contains empty topic id")
		}

		value.Name = strings.TrimSpace(value.Name)
		if value.Name == "" {
			return fmt.Errorf("bus.topics.%s.name is required", key)
		}
		if value.Partitions <= 0 {
			return fmt.Errorf("bus.topics.%s.partitions must be greater than 0", key)
		}
		if value.ReplicationFactor <= 0 {
			return fmt.Errorf("bus.topics.%s.replication_factor must be greater than 0", key)
		}

		topics[topicID] = topic(value)
	}
	c.topics = topics

	if _, ok := c.topics[TopicDefault]; !ok {
		return fmt.Errorf("bus.topics.%s is required", TopicDefault)
	}

	for key, consumer := range c.Consumers {
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("bus.consumers contains empty key")
		}
		if len(consumer.Topics) == 0 {
			return fmt.Errorf("bus.consumers.%s.topics is required", key)
		}
		for _, topicID := range consumer.Topics {
			if _, ok := c.topics[topicID]; !ok {
				return fmt.Errorf("bus.consumers.%s.topics contains undefined topic id: %s", key, topicID)
			}
		}
	}

	return nil
}

func (c *Config) TopicName(topicID TopicID) (string, error) {
	if c == nil {
		return "", fmt.Errorf("bus config is nil")
	}

	topicID = TopicID(strings.TrimSpace(string(topicID)))
	if topicID == "" {
		return "", fmt.Errorf("topic id is required")
	}

	t, ok := c.topics[topicID]
	if !ok {
		return "", fmt.Errorf("undefined bus topic id: %s", topicID)
	}

	return t.Name, nil
}

func (c *Config) ConsumerTopicNames(consumer string) ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("bus config is nil")
	}

	consumer = strings.TrimSpace(consumer)
	if consumer == "" {
		return nil, fmt.Errorf("consumer id is required")
	}

	consumerCfg, ok := c.Consumers[consumer]
	if !ok {
		return nil, fmt.Errorf("undefined bus consumer: %s", consumer)
	}

	rawTopics := make([]string, 0, len(consumerCfg.Topics))
	for _, topicID := range consumerCfg.Topics {
		topicName, err := c.TopicName(topicID)
		if err != nil {
			return nil, err
		}
		rawTopics = append(rawTopics, topicName)
	}

	return rawTopics, nil
}

func (c *Config) TopicsForProvision() []topic {
	if c == nil {
		return nil
	}

	ids := make([]string, 0, len(c.topics))
	for topicID := range c.topics {
		ids = append(ids, string(topicID))
	}
	sort.Strings(ids)

	topics := make([]topic, 0, len(ids))
	for _, id := range ids {
		topics = append(topics, c.topics[TopicID(id)])
	}

	return topics
}
