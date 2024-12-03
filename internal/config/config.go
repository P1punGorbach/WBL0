package config

type Config struct{
	Kafka Kafka `json:"kafka"`
	Pg Pg `json:"pg"`
}