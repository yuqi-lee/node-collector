package main

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cilium/ebpf"
	"github.com/go-ping/ping"
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

	p.Interval = time.Duration(collectorConfig.PingInterval) * time.Millisecond

	p.OnRecv = func(pkt *ping.Packet) {
		offset := (time.Now().UnixMilli() % collectorConfig.PromInterval) / collectorConfig.PingInterval
		pingRTT.WithLabelValues(collectorConfig.LocalIP, p.Addr(), strconv.FormatInt(offset, 10)).Set(float64(pkt.Rtt.Microseconds()))
	}

	return nil
}

func recordTCPConnections(podName string, ip string) error {

	offset := (time.Now().UnixMilli() % collectorConfig.PromInterval) / collectorConfig.TcpRecordInterval

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
		cmdStr := "kubectl --kubeconfig " + collectorConfig.KubeConfigPath + " exec -it -n " + collectorConfig.K8sNamespace + " " + podName + " -- wc -l /proc/net/tcp"
		//log.Printf("current command is: %s\n", cmdStr)

		cmd := exec.Command("bash", "-c", cmdStr)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("exec word count in pods(name:%s, ip:%s) falied.\n", podName, ip)
			return err
		} else {
			outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
			outStr = CutString(outStr)                     // 把字符串后面的文件名截去
			outStr = strings.Replace(outStr, "\n", "", -1) // 删去换行符，要不然转float会出错
			float64Res, err := strconv.ParseFloat(outStr, 64)
			if err != nil {
				log.Printf("recordTCPConnections: string to float64 error: %s", err.Error())
			}
			//log.Printf("pod %s has %v tcp connections", podName, float64Res)
			tcpConnections.WithLabelValues(ip, strconv.FormatInt(offset, 10)).Set(float64Res)
		}
	}

	return nil
}

func recordVethDropped(pod string, veth string) error {
	offset := (time.Now().UnixMilli() % collectorConfig.PromInterval) / collectorConfig.VethRecordInterval

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

func recordVethDroppedV2() error { //只记录 eno1 的网卡队列信息

	offset := (time.Now().UnixMilli() % collectorConfig.PromInterval) / collectorConfig.VethRecordInterval

	cmdStr := "cat /proc/net/dev | awk '/eno1:/{print $5}'"
	cmd := exec.Command("bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New("record veth dropped num failed." + err.Error())
	}

	outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
	outStr = strings.Replace(outStr, "\n", "", -1) // 删去换行符，要不然转float会出错
	float64Res, err := strconv.ParseFloat(outStr, 64)
	if err != nil {
		log.Printf("recordVethDroppedV2: string to float64 error: %s", err.Error())
	}
	//log.Printf("%s 's eno1 dropped num is %v : %s", collectorConfig.LocalIP, float64Res, outStr)
	vethDroppedNum.WithLabelValues(collectorConfig.LocalIP+":eno1", strconv.FormatInt(offset, 10)).Set(float64Res)
	return nil
}
