package types

import "time"

//FilebeatLog A filebeat log
type FilebeatLog struct {
	Timestamp time.Time `json:"@timestamp"`
	Metadata  struct {
		Beat    string `json:"beat"`
		Type    string `json:"type"`
		Version string `json:"version"`
		Topic   string `json:"topic"`
	} `json:"@metadata"`
	Offset int `json:"offset"`
	Log    struct {
		File struct {
			Path string `json:"path"`
		} `json:"file"`
	} `json:"log"`
	JSON struct {
		Log    string `json:"log"`
		Stream string    `json:"stream"`
		Time   string `json:"time"`
	} `json:"json"`
	Input struct {
		Type string `json:"type"`
	} `json:"input"`
	Beat struct {
		Hostname string `json:"hostname"`
		Version  string `json:"version"`
		Name     string `json:"name"`
	} `json:"beat"`
	Source     string `json:"source"`
	Kubernetes struct {
		Container struct {
			Name string `json:"name"`
		} `json:"container"`
		Namespace  string `json:"namespace"`
		Replicaset struct {
			Name string `json:"name"`
		} `json:"replicaset"`
		Labels struct {
			Name            string `json:"name"`
			Namespace       string `json:"namespace"`
			PodTemplateHash string `json:"pod-template-hash"`
		} `json:"labels"`
		Pod struct {
			UID  string `json:"uid"`
			Name string `json:"name"`
		} `json:"pod"`
		Node struct {
			Name string `json:"name"`
		} `json:"node"`
	} `json:"kubernetes"`
	Host struct {
		Name string `json:"name"`
	} `json:"host"`
	Prospector struct {
		Type string `json:"type"`
	} `json:"prospector"`
}