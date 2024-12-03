package config
type Kafka struct{
	Brokers []string `json:"brokers"`
	Version string `json:"version"`
	Group string `json:"group"`
	Topics string `json:"topics"`
	Assignor string `json:"assignor"`
	Oldest bool `json:"oldest"`
	Verbose string `json:"verbose"`
}