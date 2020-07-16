package main

import "github.com/Shopify/sarama"

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
    ready chan bool
    messageChan chan []byte
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for message := range claim.Messages() {
        consumer.messageChan <- message.Value
    }

    return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
    // Mark the consumer as ready
    close(consumer.ready)
    consumer.messageChan = make(chan []byte, 1000)
    go handleLogMessage(consumer.messageChan)
    return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
    close(consumer.messageChan)
    return nil
}

