package main

import "fmt"

func InetNtoA(ip int32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InttoProtocol(protocol int32) string {
	if protocol == 17 {
		return "ipv4"
	} else {
		return "ipv6"
	}
}
