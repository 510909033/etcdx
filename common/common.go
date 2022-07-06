package common

import (
	"go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func GetCliV3() *clientv3.Client {
	log.SetFlags(log.Lshortfile)
	client, err := clientv3.New(clientv3.Config{
		// 集群列表
		//Endpoints:   []string{IP},
		Endpoints: []string{IP, IP1, IP2},
		//Endpoints:   []string{IP2},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	return client
}

var IP = "172.20.10.40:2379"

var IP1 = "172.20.10.40:22379"

var IP2 = "172.20.10.40:32379"
