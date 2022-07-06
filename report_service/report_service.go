package report_service

import (
	"context"
	"encoding/json"
	"github.com/510909033/etcdx/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"os"
	"time"
)

func Report() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("出现了panic, err=%+v", err)
		}
		ticker := time.NewTicker(time.Second * 2)
		timer := time.NewTimer(time.Minute)
		for {
			select {
			case msg := <-ticker.C:
				log.Printf("%s", msg)
			case <-timer.C:
				return
			}
		}
	}()

	ipList, err := common.GetLocalIPList()
	if err != nil {
		panic(err)
	}

	name, err := os.Hostname()

	if err != nil {
		panic(err)
	}

	key := "/ipv1/" + name

	ipListBytes, err := json.Marshal(ipList)
	if err != nil {
		panic(err)
	}
	val := string(ipListBytes)

	client := common.GetCliV3()
	ctx := context.Background()

	clientv3.WithPrefix()

	opts := []clientv3.OpOption{
		clientv3.WithPrevKV(),
	}
	response, err := client.Put(ctx, key, val, opts...)
	if err != nil {
		log.Printf("put err=%+v", err)
		panic(err)
	} else {
		if response.PrevKv != nil {
			log.Printf("PrevKv.value=%s", response.PrevKv.Value)
		}
		if response.OpResponse().Get() != nil {
			log.Printf("response.OpResponse().Get().Kvs[0].Value=%s", response.OpResponse().Get().Kvs[0].Value)
		}
	}

	log.Println("over")
	//time.Sleep(time.Second * 3)
}
