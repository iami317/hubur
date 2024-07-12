package hubur

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

func GetLocalIp() (ips []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To16() != nil {
				if !strings.Contains(ipnet.IP.String(), "172.") &&
					!strings.Contains(ipnet.IP.String(), "10.") &&
					!strings.Contains(ipnet.IP.String(), "169.") &&
					!strings.Contains(ipnet.IP.String(), "fe80:") {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	return ips, nil
}

func IpRange(start, end string) (ips []string, err error) {
	startIP := net.ParseIP(start)
	endIP := net.ParseIP(end)
	if startIP.To16() == nil || endIP.To16() == nil || !startIP.To16().Equal(startIP) || !endIP.To16().Equal(endIP) {
		return nil, fmt.Errorf("Invalid ip addresses")
	}

	for ip := startIP; bytes.Compare(startIP, endIP) <= 0; incrementIPv6(ip) {
		ips = append(ips, ip.To16().String())
	}
	return
}

// Function to increment an IPv6 address by 1
func incrementIPv6(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}
