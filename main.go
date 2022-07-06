package main

import (
	"context"
	"github.com/510909033/etcdx/report_service"
)

var (
	ctx = context.TODO()
)

func main() {

	//discoveryService := discovery.NewDiscoveryService()
	//discoveryService.DemoMultiHttpAndRegister(common.GetCliV3(), "/ip_list")

	report_service.Report()

	//cli, err := clientv3.New(clientv3.Config{
	//	Endpoints:   []string{common.IP1, common.IP1, common.IP2},
	//	DialTimeout: 5 * time.Second,
	//})
	//if err != nil {
	//	panic(err)
	//	// handle error!
	//}
	//defer cli.Close()
}
