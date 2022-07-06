// 服务发现的示例
package discovery

import (
	"context"
	"fmt"
	"github.com/510909033/etcdx/common"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"net/http"
	"sync"
	"time"
)

type discoveryService struct {
	mu   sync.Mutex
	Port int64
	Ip   string
}

func NewDiscoveryService() *discoveryService {
	return &discoveryService{}
}

//172.16.7.228
func (service *discoveryService) GetIP() string {
	//todo
	return "127.0.0.1"
	ip, err := common.GetLocalIP()
	if err != nil {
		log.Printf("GetLocalIP err=%+v\n", err)
		return "6.6.6.6"
	}
	return ip
}

func (service *discoveryService) GetPort() int64 {
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.Port < 1 {
		service.Port = 33333
	}
	service.Port++
	return service.Port
}

/*
1. 监听前缀为discovery的key
*/
func (service *discoveryService) watch(client *clientv3.Client, key string) {

	ctx := context.Background()

	watcher := clientv3.NewWatcher(client)
	defer func() {
		log.Println("执行defer方法")
		err := watcher.Close()
		if err != nil {
			log.Printf("watcher.Close出错， err=%+v", err)
		} else {
			log.Println("watcher.Close成功")
		}
	}()
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
	}
	watchChan := watcher.Watch(ctx, key, opts...)
	log.Println("开始遍历watchChan")
	for val := range watchChan {
		log.Println("收到chan消息")
		//log.Printf("%#v", val.Created, val.Canceled, val.CompactRevision, val.Err(), val.IsProgressNotify())
		log.Printf("事件内容为： %#v", val)
		for _, event := range val.Events {
			log.Printf("event.Type=%s", event.Type)
			log.Printf("event.Kv=%+v", event.Kv)
			log.Printf("event.PrevKv=%+v", event.PrevKv)
			log.Printf("event.IsCreate=%t", event.IsCreate())
			log.Printf("event.IsModify=%t", event.IsModify())
		}
	}

}

// 每秒钟put一次 key的值
//可以用来打印etcd每次的结果，看看返回值是什么
//
//可以设置opts来实验不同的功能
func (service *discoveryService) DemoPut(client *clientv3.Client, key string) {
	ctx := context.Background()
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			opts := []clientv3.OpOption{
				clientv3.WithPrevKV(),
			}
			response, err := client.Put(ctx, key, time.Now().String(), opts...)
			if err != nil {
				log.Printf("put err=%+v", err)
			} else {
				if response.PrevKv != nil {
					log.Printf("PrevKv.value=%s", response.PrevKv.Value)
				}
				if response.OpResponse().Get() != nil {
					log.Printf("response.OpResponse().Get().Kvs[0].Value=%s", response.OpResponse().Get().Kvs[0].Value)
				}

			}
		}
	}

}

// 生成多个http服务，并将每一个服务的ip:port注册到 key的前缀中
//
//同时开启一个协程,watch key的变化
//
func (service *discoveryService) DemoMultiHttpAndRegister(client *clientv3.Client, key string) {

	go func() {
		util := ToolsUtil{}
		for {
			count := util.GetRegisterCount(context.Background(), client, key)
			logrus.Infof("GetRegisterCount = %d", count)
			time.Sleep(time.Second * 5)
		}
	}()

	go service.watch(client, key)
	time.Sleep(time.Second * 3)

	go service.demoHttp(client, key)
	//for i := 0; i < 2; i++ {
	//	go service.demoHttp(client, key)
	//}
	select {}
}

//启动一个http服务，端口service.GetPort() 累加获取，每次不同
//将addr注册到etcd
func (service *discoveryService) demoHttp(client *clientv3.Client, key string) {
	addr := fmt.Sprintf("%s:%d", service.GetIP(), service.GetPort())

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(addr))
	}
	go func() {
		for {
			resp, err := http.Get("http://" + addr + "/")
			if err != nil {
				log.Println(addr + " 服务尚未启动")
				time.Sleep(time.Second)
				continue
			}
			log.Println(addr+" 服务已启动， 结果为", resp)
			log.Println(addr + " 开始服务注册")
			service.register(client, key, addr)
		}
	}()
	panic(http.ListenAndServe(addr, http.HandlerFunc(handler)))
}
func (service *discoveryService) register(client *clientv3.Client, key string, val string) {
	ctx := context.Background()
	ctx, cancle := context.WithTimeout(ctx, time.Second*5)
	defer cancle()

	ticker := time.NewTicker(time.Second)
	key = key + "/" + val
	log.Println("注册的key为：", key)
	for {
		select {
		case <-ticker.C:

			grantResponse, err := client.Grant(ctx, 8)
			if err != nil {
				log.Printf("ERROR ,Grant fail, err=%+v\n", err)
				break
			}
			log.Printf("DEBUG grantResponse.LeaseId=%d\n", grantResponse.ID)

			opts := []clientv3.OpOption{
				clientv3.WithPrevKV(),
				clientv3.WithLease(grantResponse.ID),
			}
			response, err := client.Put(ctx, key, val, opts...)
			if err != nil {
				log.Printf("put err=%+v", err)
			} else {
				if response.PrevKv != nil {
					log.Printf("PrevKv.value=%s", response.PrevKv.Value)
				}
				if response.OpResponse().Get() != nil {
					log.Printf("response.OpResponse().Get().Kvs[0].Value=%s", response.OpResponse().Get().Kvs[0].Value)
				}

				//续约
				for { //这个for应该是用不到的
					aliveResponses, err := client.KeepAlive(context.Background(), grantResponse.ID)
					if err != nil {
						log.Printf("ERROR KeepAlive err=%+v\n", err)
					} else {
						for ar := range aliveResponses {
							log.Printf("recive aliveResponses id=%d, ttl=%d\n", ar.ID, ar.TTL)
						}
					}
					time.Sleep(time.Millisecond * 10)
				}
			} //end else
		}
	}

}
