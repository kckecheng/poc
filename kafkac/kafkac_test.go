package kafkac

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
)

var addrs []string = []string{"localhost:9092"}

func init() {
	bs, ok := os.LookupEnv("KAFKA_BROKERS")
	if ok {
		addrs = strings.Split(bs, ",")
	}
}

func makeConn() *kafkaConn {
	c, err := NewC(addrs)
	if err != nil {
		panic(err)
	}
	return c
}

func TestInitConfig(t *testing.T) {
	_ = initConfig()
	// t.Logf("Init config: %#v", *config)
}

func TestNewC(t *testing.T) {
	c := makeConn()
	defer c.Close()
}

func TestSendMessage(t *testing.T) {
	c := makeConn()
	defer c.Close()

	for i := 0; i < 10; i++ {
		err := c.SendMessage("kafkac", "TestSendMessage", fmt.Sprintf("This is message %v", i))
		if err != nil {
			t.Log(err)
		}
	}
}

func TestConsume(t *testing.T) {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	c := makeConn()
	defer c.Close()

	msgch, err := c.Consume("kafkac", "localhost")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Press Ctr + C to terminal this test")
	for {
		select {
		case msg := <-msgch:
			t.Logf("Topic: %v, Partition: %v, Key: %v, Offset: %v. Value: %s", msg.Topic, msg.Partition, msg.Key, msg.Offset, msg.Value)
		case <-sigterm:
			t.Log("Test done")
			return
		}
	}
}
