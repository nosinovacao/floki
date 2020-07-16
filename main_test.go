package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	p "github.com/nosinovacao/floki/logproto"
	types "github.com/nosinovacao/floki/types"

	"github.com/golang/snappy"
)

func logMsg(msg interface{}, t *testing.T) {
	fmt.Println(msg)
	t.Log(msg)
}


func getFakeLog(rightnow time.Time) *types.FilebeatLog {
	log := &types.FilebeatLog{}
	log.Timestamp = rightnow
	log.Metadata.Beat = "beat"
	log.Metadata.Type = "type"
	log.Metadata.Version = "1"
	log.Metadata.Topic = "topic"
	log.Offset = 0
	log.Log.File.Path = "path"
	log.JSON.Stream = "stream"
	log.JSON.Time = rightnow.Format(time.RFC3339)
	log.JSON.Log = "somelog"
	log.Input.Type = "inputtype"
	log.Beat.Hostname = "hostname"
	log.Beat.Version = "1"
	log.Beat.Name = "name"
	log.Source = "source"
	log.Kubernetes.Container.Name = "containername"
	log.Kubernetes.Namespace = "namespace"
	log.Kubernetes.Replicaset.Name = "replicaset"
	log.Kubernetes.Labels.Name = "name"
	log.Kubernetes.Labels.Namespace = "namespace"
	log.Kubernetes.Labels.PodTemplateHash = "hash"
	log.Kubernetes.Pod.UID = "12312211aa"
	log.Kubernetes.Pod.Name = "pod"
	return log
}

func TestLokiLogTransfer(t *testing.T) {
	rightnow := time.Now()

	testserver := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		bytes, err := ioutil.ReadAll(req.Body)

		if err != nil {
			t.Fatalf("unable to parse body: %v", err)
			return
		}

		bytes, err = snappy.Decode(nil, bytes)

		if err != nil {
			t.Fatalf("unable to snappy decode body: %v", err)
			return
		}

		var body p.PushRequest
		err = body.Unmarshal(bytes)

		if err != nil {
			t.Errorf("Unable to unmarshall protobuf to PushRequest: %v", err)
			return
		}

		if len(body.Streams) != 1 {
			t.Errorf("Expected just one stream, got sent %d. %v", len(body.Streams), body)
			return
		}

		log := body.Streams[0]

		if len(log.Entries) != 1 {
			t.Errorf("Got sent more than one log: %v", len(log.Entries))
		}

		entry := log.Entries[0]

		if entry.Timestamp.Unix() != rightnow.Unix() || entry.Line != "somelog" {
			t.Errorf("Unexpected log: %v %v", entry.Timestamp, entry.Line)
		}

		res.WriteHeader(201)
		res.Write([]byte("test body"))
	}))

	defer testserver.Close()

	lokiURL = testserver.URL

	var col []*types.FilebeatLog

	log := getFakeLog(rightnow)

	col = append(col, log)

	sendToLoki(col)
}

func TestLogHandlingFromChannel(t *testing.T) {
	ch := make(chan []byte)
	rightnow := time.Now()
	duration := 5 * time.Second
	internalBuffer = &duration

	logMsg("starting handleLogMessage", t)
	go handleLogMessage(ch)

	logMsg("creating and sending fake log", t)
	log := getFakeLog(rightnow)
	logJSON, err := json.Marshal(log)

	if err != nil {
		t.Fatalf("Cannot marshall log to json: %v", err)
		return
	}

	logMsg("sending logJSON", t)
	ch <- logJSON

	logMsg("testing `collection` length", t)
	// collection now should have a log
	if len(collection) != 1 {
		t.Errorf("`collection` should have one log. len(collection)=%d", len(collection))
		return
	}

	// let the buffer time pass
	time.Sleep(5 * time.Second)
}
