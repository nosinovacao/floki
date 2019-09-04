package main

import (
	"testing"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func TestKafkaConsumerCreationEmpty(t *testing.T) {
	consumer, err := kafkaConsumer(&kafka.ConfigMap{})

	if consumer != nil || err == nil {
		t.Error("Should not be possible to create a consumer with empty parameters")
	}

	t.Logf("Could not create consumer with empty parameters: %v", err)
}

func TestKafkaConsumerCreationSomeParameters(t *testing.T) {
	consumer, err := kafkaConsumer(&kafka.ConfigMap{
		"group.id": "someid",
	})

	if err != nil {
		t.Errorf("Should have been able to create a consumer: %v", err)
	}

	t.Logf("Consumer created: %v", consumer)
}
