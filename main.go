package main

import (
    "context"
    "encoding/json"
    "flag"
    "github.com/Shopify/sarama"
    "github.com/hashicorp/go-uuid"
    "os"
    "os/signal"
    "strings"
    "sync"
    "syscall"
    "time"

    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/log/level"

    t "github.com/nosinovacao/floki/types"
)

var (
	logger = log.With(log.NewJSONLogger(os.Stdout), "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	debug  = level.Debug(logger)
	info   = level.Info(logger)
	errorl = level.Error(logger)

	lokiURL        string
	internalBuffer *time.Duration
)


func handleLogMessage(ch chan []byte) {
	ticker := time.NewTicker(*internalBuffer)
	defer ticker.Stop()
	for {
		select {
		case data := <-ch:
			obj := &t.FilebeatLog{}
			err := json.Unmarshal(data, obj)
			if err != nil {
				errorl.Log("msg", "unable to unmarshall json data", "err", err)
			} else {
				//debug.Log("ts", obj.Timestamp, "instance", obj.Host.Name, "ns", obj.Kubernetes.Namespace, "pod", obj.Kubernetes.Pod.Name, "msg", obj.JSON.Log)
				collection = append(collection, obj)
			}
		case <-ticker.C:
		    collectionLength := len(collection)
			if collectionLength == 0 {
				continue
			}
			debug.Log("msg", "sorting and processing", "collectionLength", collectionLength)
			copyOfCollection := make(logCollection, collectionLength)
			copy(copyOfCollection, collection)
			go sendToLoki(copyOfCollection)
			collection = make(logCollection, 0, 1000)
		}
	}
}

func main() {

	var (
		lokiAddr       = flag.String("lokiurl", "http://loki:3100", "the loki url, default is http://loki:3100")
		buffer         = flag.Duration("buffer", 5*time.Second, "how much time to buffer before sending to loki")
		brokerList     = flag.String("brokerList", "awesome.kafka.broker:32400", "the kafka broker list")
		clientID       = flag.String("clientid", "", "the kafka client id")
		groupID        = flag.String("groupid", "floki", "the kafka group id")
		topicPatterns  = flag.String("topicPatterns", "^logging-*", "Regex pattern for kafka topic subscription")
	)

	flag.Parse()

	lokiURL = *lokiAddr
	internalBuffer = buffer

	sarama.Logger = saramaLogger{}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.MaxWaitTime = time.Hour
	config.Net.MaxOpenRequests = 100
	config.Net.DialTimeout = 10
	config.Net.ReadTimeout = 5 * time.Second
	config.Net.WriteTimeout = config.Net.ReadTimeout

	if *clientID != "" {
	    config.ClientID = *clientID
    } else {
        if hostname, err := os.Hostname(); err == nil {
            config.ClientID = hostname
        } else if id, err := uuid.GenerateUUID(); err == nil {
            config.ClientID = id
        } else {
            config.ClientID = "floki"
        }
    }

    consumer := Consumer{
        ready: make(chan bool),
        messageChan: make(chan []byte, 1000),
    }

    ctx, cancel := context.WithCancel(context.Background())
    client, err := sarama.NewConsumerGroup(strings.Split(*brokerList, ","), *groupID, config)
    if err != nil {
        errorl.Log("msg", "Error creating consumer group client", "err", err)
        os.Exit(9)
    }
    wg := &sync.WaitGroup{}
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            // `Consume` should be called inside an infinite loop, when a
            // server-side rebalance happens, the consumer session will need to be
            // recreated to get the new claims
            if err := client.Consume(ctx, strings.Split(*topicPatterns, ","), &consumer); err != nil {
                errorl.Log("msg", "Error from consumer", "err", err)
                os.Exit(8)
            }
            // check if context was cancelled, signaling that the consumer should stop
            if ctx.Err() != nil {
                return
            }
            consumer.ready = make(chan bool)
        }
    }()

    <-consumer.ready // Await till the consumer has been set up
    info.Log("msg", "Sarama consumer up and running!...")

    sigterm := make(chan os.Signal, 1)
    signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
    select {
    case <-ctx.Done():
        info.Log("msg", "terminating: context cancelled")
    case <-sigterm:
        info.Log("msg", "terminating: via signal")
    }
    cancel()
    wg.Wait()
    if err = client.Close(); err != nil {
        errorl.Log("msg", "Error closing client", "err", err)
        os.Exit(6)
    }
}
