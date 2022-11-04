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
	}, []string{"src", "dst"})
)

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
				fmt.Printf("monitor bytes and packets failed: %s", err.Error())
			}
			cnt = cnt + 1
			fmt.Printf("Round %d is finished.", cnt)
		}
	}()
}

func monitorPingRTT() {
	go func() {
		err := recordPingRTT("192.168.1.116")
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := recordPingRTT("192.168.1.103")
		if err != nil {
			panic(err)
		}
	}()
}

func main() {
	//bpfMapLoad()

	monitorBytesAndPackets()

	monitorPingRTT()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":40901", nil)
}
