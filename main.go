package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/snappy"
	l "github.com/nosinovacao/floki/logproto"
	t "github.com/nosinovacao/floki/types"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	logger = log.With(log.NewJSONLogger(os.Stdout), "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	debug  = level.Debug(logger)
	info   = level.Info(logger)
	errorl = level.Error(logger)

	lokiURL        string
	internalBuffer *time.Duration
)

type logCollection []*t.FilebeatLog

var collection = make(logCollection, 0, 1000)

func sendToGrafana(col logCollection) {
	debug.Log("msg", "sendToGrafana", "len", len(col))

	body := map[string][]l.Entry{}

	for _, filebeatLog := range col {
		labels := fmt.Sprintf("{name=\"%s\", namespace=\"%s\", instance=\"%s\", replicaset=\"%s\"}",
			filebeatLog.Kubernetes.Container.Name,
			filebeatLog.Kubernetes.Namespace,
			filebeatLog.Kubernetes.Node.Name,
			filebeatLog.Kubernetes.Replicaset.Name)
		entry := l.Entry{Timestamp: filebeatLog.Timestamp, Line: filebeatLog.JSON.Log}
		if val, ok := body[labels]; ok {
			body[labels] = append(val, entry)
		} else {
			body[labels] = []l.Entry{entry}
		}
	}

	streams := make([]*l.Stream, 0, len(body))

	for key, val := range body {
		stream := &l.Stream{
			Labels:  key,
			Entries: val,
		}
		sort.SliceStable(stream.Entries, func(i, j int) bool {
			return stream.Entries[i].Timestamp.Before(stream.Entries[j].Timestamp)
		})
		streams = append(streams, stream)
	}

	debug.Log("msg", fmt.Sprintf("sending %d streams to the server", len(streams)))
	req := l.PushRequest{
		Streams: streams,
	}

	buf, err := req.Marshal() //proto.Marshal(&req)

	if err != nil {
		errorl.Log("msg", "unable to marshall PushRequest", "err", err)
		return
	}

	buf = snappy.Encode(nil, buf)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	httpreq, err := http.NewRequest(http.MethodPost, lokiURL, bytes.NewReader(buf))

	if err != nil {
		errorl.Log("msg", "unable to create http request", "err", err)
		return
	}

	httpreq = httpreq.WithContext(ctx)
	httpreq.Header.Add("Content-Type", "application/x-protobuf")

	client := http.Client{}

	resp, err := client.Do(httpreq)

	if err != nil {
		errorl.Log("msg", "http request error", "err", err)
		return
	}

	if resp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, 1024))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s (%d): %s", resp.Status, resp.StatusCode, line)
		errorl.Log("mng", "server returned an error", "err", err)
		return
	}
	info.Log("msg", "sent streams to loki")
}

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
			if len(collection) == 0 {
				continue
			}
			debug.Log("msg", "sorting and processing", "len", len(collection))
			debug.Log("msg", "going to process messages", "len", len(collection))
			go sendToGrafana(collection)
			collection = make(logCollection, 0, 1000)
		}
	}
}

func kafkaConsumer(cfg *kafka.ConfigMap) (*kafka.Consumer, error) {
	c, err := kafka.NewConsumer(cfg)
	return c, err
}

func main() {

	var (
		lokiAddr       = flag.String("lokiurl", "http://loki:3100", "the loki url, default is http://loki:3100")
		buffer         = flag.Duration("buffer", 5*time.Second, "how much time to buffer before sending to loki")
		brokerList     = flag.String("brokerList", "awesome.kafka.broker:32400", "the kafka broker list")
		clientID       = flag.String("clientid", "floki", "the kafka client id")
		groupID        = flag.String("groupid", "floki", "the kafka group id")
		sessionTimeout = flag.String("sessionTimeout", "6000", "kafka session timeout")
		topicPattern   = flag.String("topicPattern", "^logging-*", "Regex pattern for kafka topic subscription")
	)
	flag.Parse()

	lokiURL = *lokiAddr
	internalBuffer = buffer

	c, err := kafkaConsumer(&kafka.ConfigMap{
		"metadata.broker.list": *brokerList,
		"client.id":            *clientID,
		"group.id":             *groupID,
		"session.timeout.ms":   *sessionTimeout,
		"auto.offset.reset":    "latest",
	})

	if err != nil {
		fmt.Printf("Unable to create consumer: %v", err)
		os.Exit(1)
	}
	info.Log("msg", "consumer created.")

	defer c.Close()

	err = c.SubscribeTopics([]string{*topicPattern}, nil)

	if err != nil {
		fmt.Printf("Unable to subscribe to subscribe to topics: %v", err)
		os.Exit(2)
	}

	info.Log("msg", "topic subscribed.")

	ch := make(chan []byte, 1000)
	go handleLogMessage(ch)
	defer close(ch)

	for {
		msg, err := c.ReadMessage(1 * time.Minute)
		if err == nil {
			ch <- msg.Value
		} else {
			errorl.Log("msg", "Consumer error", "err", err, "kafkamsg", msg)
			os.Exit(-9)
		}
	}
}
