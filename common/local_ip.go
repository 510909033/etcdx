package common

import (
	"log"
	"net"
)

//http://ldaysjun.com/2019/01/14/Microservice/micro3/
func GetLocalIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	panic("unable to determine locla ip")
}

func GetLocalIPList() (ipList []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipList = append(ipList, ipnet.IP.String())
			}
		}
	}
	return ipList, nil
}
