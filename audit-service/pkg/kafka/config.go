package kafka

type Config struct {
	Kafka struct {
		Brokers []string `yaml:"brokers"`
		GroupID string   `yaml:"group_id"`
		Topics  []string `yaml:"topics"`
	} `yaml:"kafka"`
}
