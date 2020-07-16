package main

import (
    "bufio"
    "bytes"
    "context"
    "fmt"
    "github.com/golang/snappy"
    l "github.com/nosinovacao/floki/logproto"
    t "github.com/nosinovacao/floki/types"
    "io"
    "net/http"
    "sort"
    "time"
)

type logCollection []*t.FilebeatLog

var collection = make(logCollection, 0, 1000)

func pushToLoki(streams []*l.Stream) {
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

func sendToLoki(col logCollection) {
    debug.Log("msg", "sendToLoki", "len", len(col))

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
    pushToLoki(streams)
}
