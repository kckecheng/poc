package kafkac

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
)

type kafkaConn struct {
	addresses []string // brokers
	client    sarama.Client
	producer  sarama.SyncProducer
}

func initConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	version, _ := sarama.ParseKafkaVersion("2.6.0") // Hard coded: to be enhanced
	config.Version = version
	return config
}

// NewC init a Kafka client
func NewC(addrs []string) (*kafkaConn, error) {
	config := initConfig()
	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return &kafkaConn{addrs, client, producer}, nil
}

// Close disconnect Kafka connection
func (c *kafkaConn) Close() error {
	c.producer.Close()
	return c.client.Close()
}

// SendMessage send message
func (c *kafkaConn) SendMessage(topic, key, message string) error {
	msg := sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	}
	_, _, err := c.producer.SendMessage(&msg)
	return err
}

// Kafka consumer group handler
type cgHandler struct {
	message chan *sarama.ConsumerMessage
}

func (h *cgHandler) Setup(sarama.ConsumerGroupSession) error {
	fmt.Println("Consumer group setup done")
	return nil
}

func (h *cgHandler) Cleanup(sarama.ConsumerGroupSession) error {
	fmt.Println("Consumer group clean up done")
	return nil
}

func (h *cgHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.message <- msg
		sess.MarkMessage(msg, "")
	}
	return nil
}

// Consume consume message
func (c *kafkaConn) Consume(topic, group string) (chan *sarama.ConsumerMessage, error) {
	cg, err := sarama.NewConsumerGroupFromClient(group, c.client)
	if err != nil {
		return nil, err
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// print out without taking any other actions
	go func() {
		for err := range cg.Errors() {
			log.Printf("Hit Kafka error: %v", err)
		}
	}()

	cgh := cgHandler{
		message: make(chan *sarama.ConsumerMessage),
	}
	ctx := context.Background()
	go func() {
		defer cg.Close()
		defer close(cgh.message)
		for {
			select {
			case <-sigterm:
				return
			default:
				err := cg.Consume(ctx, []string{topic}, &cgh)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
	return cgh.message, nil
}
