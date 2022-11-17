package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cilium/ebpf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	bytesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bytes_total",
		Help: "The total number of bytes from ip_src to ip_dst",
	}, []string{"src", "dst", "protocol"})

	packetsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "packets_total",
		Help: "The total number of packets from ip_src to ip_dst",
	}, []string{"src", "dst", "protocol"})

	pingRTT = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ping_rtt_time",
		Help: "ping rtt time from ip_src to ip_dst (us)",
	}, []string{"src", "dst", "offset"})

	tcpConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcp_connections",
		Help: "number of spod's tcp connections",
	}, []string{"pod", "offset"})

	vethDroppedNum = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "veth_dropped_num",
		Help: "number of (pod's) veth's tx_dropped",
	}, []string{"pod", "offset"})
)

func collectorInit() {
	CollectorConfigInit()
	podIPInfoInit()
	// podVethInfoInit() 暂时只统计 eno1
}

func monitorBytesAndPackets() {
	go func() {
		mp, err := ebpf.LoadPinnedMap("/sys/fs/bpf/try", nil)
		if err != nil {
			panic(err)
		} else {
			log.Printf("bpf map is loaded successfully with key size %d bytes.", mp.KeySize())
		}
		cnt := 0

		for {
			time.Sleep(time.Second)
			err := recordBytesAndPacketsTotal(mp)
			if !errors.Is(err, ebpf.ErrKeyNotExist) {
				fmt.Printf("monitor bytes and packets failed: %s\n", err.Error())
			}
			cnt = cnt + 1
			log.Printf("ebpf map reading: round %d is finished.", cnt)
		}
	}()
}

func monitorPingRTT() {
	go func() {
		err := recordPingRTT(collectorConfig.TargetIP1)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := recordPingRTT(collectorConfig.TargetIP2)
		if err != nil {
			panic(err)
		}
	}()
}

func monitorTCPConnections() {
	go func() {
		for {
			time.Sleep(time.Duration(collectorConfig.TcpRecordInterval) * time.Millisecond)

			mapName2ip.Range(func(key, value interface{}) bool {
				name := key.(string)
				ip := value.(string)

				ipaddr := net.ParseIP(ip)

				if name != "" && ipaddr != nil {
					go func() {
						err := recordTCPConnections(name, ipaddr.String())
						if err != nil {
							log.Printf("name:%s, ip:%s record tcp connections failed:%s\n", name, ipaddr.String(), err.Error())
						}
					}()
				}

				return true
			})
		}
	}()

	updateName2ip()
}

func monitorVethDroppedNum() {
	go func() {
		for {
			time.Sleep(time.Duration(collectorConfig.VethRecordInterval) * time.Millisecond)

			err := recordVethDroppedV2()
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

func main() {
	collectorInit()

	monitorBytesAndPackets()
	monitorPingRTT()

	if collectorConfig.TcpRecordInterval > 0 {
		monitorTCPConnections()
	}

	if collectorConfig.VethRecordInterval > 0 {
		monitorVethDroppedNum()
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":40901", nil)
}
