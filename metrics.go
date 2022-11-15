package main

import (
	"errors"
	"os/exec"
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

	tcpRecordInterval  = 100 // 读取tcp连接数的时间间隔毫秒数
	vethRecordInterval = 100 // 读取网卡丢包数据的时间间隔毫秒数

	k8sNamespace string = "hotel-reservation"
)

func recordBytesAndPacketsTotal(mp *ebpf.Map) error {
	var key1, key2 Key
	var value Value
	var err error

	err = mp.NextKey(key1, &key2)

	// 用 key 遍历 bpf map，结束时返回 ebpf.ErrKeyNotExist
	for err == nil {
		//fmt.Println(key2)
		mp.Lookup(key2, &value)
		//fmt.Println(value)

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

func recordTCPConnections(podName string, ip string) error {

	offset := (time.Now().UnixMilli() % promInterval) / tcpRecordInterval

	if podName == "tcp-test" {
		// 测试用，上传host本地的总TCP连接数
		cmd := exec.Command("wc", "-l", "/proc/net/tcp")
		err := cmd.Run()
		if err != nil {
			return err
		} else {
			rawRes, _ := cmd.CombinedOutput()
			float64Res, _ := strconv.ParseFloat(CutString(string(rawRes)), 64)

			tcpConnections.WithLabelValues("host-skvnode4", strconv.FormatInt(offset, 10)).Set(float64Res)
		}
	} else {
		cmdStr := "kubectl exec -it -n" + k8sNamespace + podName + "-- wc -l /proc/net/tcp"
		cmd := exec.Command("bash", "-c", cmdStr)
		err := cmd.Run()
		if err != nil {
			return err
		} else {
			rawRes, _ := cmd.CombinedOutput()
			float64Res, _ := strconv.ParseFloat(CutString(string(rawRes)), 64)

			tcpConnections.WithLabelValues("host-skvnode4", strconv.FormatInt(offset, 10)).Set(float64Res)
		}
	}

	return nil
}

func recordVethDropped(pod string, veth string) error {
	offset := (time.Now().UnixMilli() % promInterval) / vethRecordInterval

	catPath := "/sys/class/net/" + veth + "/statistics/tx_dropped"
	cmd := exec.Command("cat", catPath)
	err := cmd.Run()
	if err != nil {
		return errors.New("record veth dropped num failed." + err.Error())
	}

	rawRes, _ := cmd.CombinedOutput()
	float64Res, _ := strconv.ParseFloat(string(rawRes), 64)

	vethDroppedNum.WithLabelValues(pod, strconv.FormatInt(offset, 10)).Set(float64Res)

	return nil
}
