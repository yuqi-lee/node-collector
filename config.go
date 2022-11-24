package main

import (
	"encoding/json"
	"errors"
	"os"
)

const configPath string = "config.json"

type CollectorConfig struct {
	LocalIP               string `json:"local_ip"`
	TargetIP1             string `json:"target_ip_1"`
	TargetIP2             string `json:"target_ip_2"`
	K8sNamespace          string `json:"k8s_namespace"`
	HostName              string `json:"host_name"`
	KubeConfigPath        string `json:"kube_config_path"`
	PromInterval          int64  `json:"prom_interval"`
	PingInterval          int64  `json:"ping_interval"`
	TcpRecordInterval     int64  `json:"tcp_record_interval"`
	VethRecordInterval    int64  `json:"veth_record_interval"`
	PackageRecordInterval int64  `json:"package_record_interval"`
	BytesRecordInterval   int64  `json:"bytes_record_interval"`
	BpfMapPath            string `json:"bpf_map_path"`
	PodInfoUpdateInterval int64  `json:"pod_info_update_interval"`
}

var collectorConfig CollectorConfig = CollectorConfig{}

func CollectorConfigInit() {
	file, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&collectorConfig)
	if err != nil {
		panic(err)
	}

	if collectorConfig.PromInterval == 0 || collectorConfig.PingInterval == 0 {
		panic(errors.New("prometheus or ping interval has not been set yet"))
	}
}
