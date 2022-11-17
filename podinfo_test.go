package main

import (
	"fmt"
	"net"
	"testing"
)

func TestGenMap(t *testing.T) {
	outStr := "redis 10.24.1.23\nmongodb 10.24.1.56\nnginx 10.24.8.10\n"

	genMap(outStr)

	cnt := 0
	mapName2ip.Range(func(key, value interface{}) bool {
		name := key.(string)
		ip := value.(string)
		ipaddr := net.ParseIP(ip)
		if ipaddr == nil {
			t.Errorf("wrong ip: %s", ip)
		}

		fmt.Printf("Name is: %s, IP is: %s\n", name, ip)
		cnt += 1
		return true
	})

	if cnt != 3 {
		t.Errorf("pod num error with %d pods", cnt)
	}
}
