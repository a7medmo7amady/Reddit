package kafka

type Config struct {
	Kafka struct {
		Brokers []string 
		GroupID string   
		Topics  []string 
	} 
}
