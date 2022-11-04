package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cilium/ebpf"
	"github.com/go-ping/ping"
)

const (
	localIP      string = "192.168.1.117"
	targetIP1    string = "192.168.1.116"
	targetIP2    string = "192.168.1.103"
	pingInterval        = 20   // pinger发包间隔毫秒数
	promInterval        = 1000 // prometheus采样间隔毫秒数
)

func recordBytesAndPacketsTotal(mp *ebpf.Map) error {
	var key1, key2 Key
	var value Value
	var err error

	err = mp.NextKey(key1, &key2)

	// 用 key 遍历 bpf map，结束时返回 ebpf.ErrKeyNotExist
	for err == nil {
		fmt.Println(key2)
		mp.Lookup(key2, &value)
		fmt.Println(value)

		strBytes := strconv.FormatInt(value.Bytes, 10)
		float64Bytes, _ := strconv.ParseFloat(strBytes, 64)
		strPackets := strconv.FormatInt(value.Packets, 10)
		float64Packets, _ := strconv.ParseFloat(strPackets, 64)

		bytesTotal.WithLabelValues(InetNtoA(key2.Src), InetNtoA(key2.Dst), InttoProtocol(key2.Protocol)).Set(float64Bytes)
		packetsTotal.WithLabelValues(InetNtoA(key2.Src), InetNtoA(key2.Dst), InttoProtocol(key2.Protocol)).Set(float64Packets)

		key1 = key2
		err = mp.NextKey(key1, &key2)
	}

	return err
}

func recordPingRTT(ip string) error {

	p, err := ping.NewPinger(ip)
	if err != nil {
		return err
	}
	defer p.Run()

	p.Interval = pingInterval * time.Millisecond

	p.OnRecv = func(pkt *ping.Packet) {
		offset := (time.Now().UnixMilli() % promInterval) / pingInterval
		pingRTT.WithLabelValues(localIP, p.Addr(), strconv.FormatInt(offset, 10)).Set(float64(pkt.Rtt.Microseconds()))
	}

	return nil
}
