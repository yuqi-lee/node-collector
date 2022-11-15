package main

import (
	"fmt"
	"log"
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
	podIPInfoInit()
	podVethInfoInit()
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
			if err != ebpf.ErrKeyNotExist {
				fmt.Printf("monitor bytes and packets failed: %s\n", err.Error())
			}
			cnt = cnt + 1
			//fmt.Printf("Round %d is finished.", cnt)
		}
	}()
}

func monitorPingRTT() {
	go func() {
		err := recordPingRTT(targetIP1)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := recordPingRTT(targetIP2)
		if err != nil {
			panic(err)
		}
	}()
}

func monitorTCPConnections() {
	go func() {
		for {
			time.Sleep(tcpRecordInterval * time.Millisecond)

			mapName2ip.Range(func(key, value interface{}) bool {
				name := key.(string)
				ip := value.(string)

				if name != "" && ip != "" {
					go func() {
						err := recordTCPConnections(name, ip)
						if err != nil {
							log.Println(err)
						}
					}()
				}

				return true
			})
		}
	}()

	//updateName2ip()
}

func monitorVethDroppedNum() {
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)

			mapName2veth.Range(func(key, value interface{}) bool {
				name := key.(string)
				veth := value.(string)

				if name != "" && veth != "" {
					go func() {
						err := recordVethDropped(name, veth)
						if err != nil {
							log.Println(err)
						}
					}()
				}

				return true
			})
		}
	}()
}

func main() {
	collectorInit()

	monitorBytesAndPackets()
	monitorPingRTT()
	monitorTCPConnections()
	// monitorVethDroppedNum()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":40901", nil)
}
