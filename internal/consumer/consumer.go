package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"zadaniel0/internal/config"
	"zadaniel0/internal/model"

	"github.com/IBM/sarama"
)

type Consumer struct {
	datachan chan model.Order
	topics   []string
	cg       sarama.ConsumerGroup
}

func NewConsumer(ctx context.Context, cfg config.Kafka, ch chan model.Order) (*Consumer, error) {
	config := sarama.NewConfig()
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("Fail to parse kafka version: %w", err)
	}
	config.Version = version
	switch cfg.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		return nil, fmt.Errorf("Fail to parse assignor: %w", err)
	}

	if cfg.Oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	cg, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Group, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		topics:   []string{cfg.Topics},
		cg:       cg,
		datachan: ch,
	}, nil
}
func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (c *Consumer) Run(ctx context.Context) error {
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.cg.Consume(ctx, c.topics, c); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}

		}
	}()
	return nil
}
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			session.MarkMessage(message, "")
			var order model.Order
			err := json.Unmarshal(message.Value, &order)
			if err != nil {
				return fmt.Errorf("Fail to unmarshal: %w", err)
			}
			c.datachan <- order
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}
