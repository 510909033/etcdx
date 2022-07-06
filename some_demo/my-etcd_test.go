package some_demo

import (
	"context"
	"github.com/510909033/etcdx/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"

	"testing"
	"time"
)

//监听 name
//etcdctl put name test_watch
//etcdctl put name test_watch
func TestEtcdWatch(t *testing.T) {
	name := "test_watch"
	log.Println("name=", name)
	cli := common.GetCliV3()

	defer func() {
		log.Println("defer")
		err := cli.Close()
		log.Println("defer, err=", err)

	}()

	// 监听值
	go func() {
		watch := cli.Watch(etcdx.ctx, name)
		for {
			res := <-watch
			t.Logf("name发生改变, res=%+v", res)
			log.Println(res.Events[0].Type)             //PUT
			log.Println(string(res.Events[0].Kv.Value)) //v3
			time.Sleep(time.Second)
		}
	}()

	select {}

}

/*
etcdctl --endpoints=$ENDPOINTS put web1 value1
etcdctl --endpoints=$ENDPOINTS put web2 value2
etcdctl --endpoints=$ENDPOINTS put web3 value3

etcdctl --endpoints=$ENDPOINTS get web --prefix
*/
//测试前缀
func TestEtcdPrefix(t *testing.T) {
	name := "v"
	log.Println("前缀测试， 前缀=", name)
	cli := common.GetCliV3()

	defer func() {
		log.Println("defer")
		err := cli.Close()
		log.Println("defer, err=", err)

	}()

	response, err := cli.Get(etcdx.ctx, name, clientv3.WithPrefix())
	if err != nil {
		log.Printf("err=%+v", err)
		return
	}
	for k, v := range response.Kvs {
		//log.Println(response.OpResponse().Get().Header)
		log.Printf("k=%d, key=%s, val=%s", k, v.Key, v.Value)
	}

}

func TestEtcdTransaction(t *testing.T) {
	log.Println("事务测试")
	cli := common.GetCliV3()

	defer func() {
		log.Println("defer")
		err := cli.Close()
		log.Println("defer, err=", err)

	}()

	//txn := cli.Txn(ctx)

	//txn.If()

	//clientv3.NewLease().KeepAliveOnce()

}

/*

etcdctl --endpoints=$ENDPOINTS lease -h
COMMANDS:
	grant		Creates leases
	keep-alive	Keeps leases alive (renew)
	list		List all active leases
	revoke		Revokes leases
	timetolive	Get lease information
*/
func TestEtcdLease_create(t *testing.T) {
	cli := common.GetCliV3()
	defer func() {
		cli.Close()
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel

	// minimum lease TTL is 5-second
	ttl := 3
	key := "foo"
	log.Println("续约", ttl, "秒时间")
	resp, err := cli.Grant(ctx, int64(ttl))
	if err != nil {
		log.Printf("resp=%+v", resp)
		log.Fatal(err)
	}
	log.Println("grant续约成功")

	keepAliveFunc := func() {
		for {
			keepAliveResponses, err := cli.KeepAlive(ctx, resp.ID)
			if err != nil {
				log.Printf("KeepAlive返回了err=%+v", err)
			}
			//如果出现了取消，等错误，这里会收到结果
			kaVal := <-keepAliveResponses
			log.Printf("chan keepAliveResponses, res=%+v", kaVal)
			break
		}
	}
	_ = keepAliveFunc

	getVal := func(key string) {
		i := 0
		for {
			response, err := cli.Get(ctx, key)
			if err != nil {
				log.Printf("获取key=%s的值失败,err=%+v", key, err)
				return
			}
			if len(response.Kvs) == 0 {
				log.Printf("key=%s的值为不存在或已过期", key)
			} else {
				log.Printf("获取key=%s的值为：%s", key, response.Kvs[0].Value)
			}
			time.Sleep(time.Millisecond * 500)
			i++
			if i > 50 {
				//cancel()
			}
		}
	}

	//go getVal(key)

	go keepAliveFunc()

	sleepTs := 3
	log.Println("暂停", sleepTs, "秒钟")
	time.Sleep(time.Second * time.Duration(sleepTs))

	// after 5 seconds, the key 'foo' will be removed
	log.Printf("开始设置key=%s的值", key)
	putResponse, err := cli.Put(context.TODO(), key, "bar", clientv3.WithLease(resp.ID))
	_ = putResponse
	if err != nil {
		log.Fatal("put失败", err)
	}
	log.Printf("设置key=%s的值成功", key)

	time.Sleep(time.Millisecond * 101)
	getVal(key)

}

//分布式锁
func TestEtcdDistributed_Lock(t *testing.T) {
	client := common.GetCliV3()
	keyPrefix := "/lock"

	// 生成一个30s超时的上下文
	timeout, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	// 获取租约
	response, e := client.Grant(timeout, 1)
	if e != nil {
		log.Fatal(e.Error())
	}
	log.Println("获取租约成功")
	// 通过租约创建session
	session, e := concurrency.NewSession(client, concurrency.WithLease(response.ID))
	if e != nil {
		log.Fatal(e.Error())
	}
	defer session.Close()
	log.Println("使用租约创建session成功")

	// 通过session和锁前缀
	mutex := concurrency.NewMutex(session, keyPrefix)
	//return &lockerMutex{NewMutex(s, pfx)}
	e = mutex.Lock(timeout)
	if e != nil {
		log.Fatal(e.Error())
	}

	// 业务逻辑
	log.Println("执行业务逻辑")
	time.Sleep(time.Second * 31)

	// 释放锁
	defer func() {
		log.Println("开始执行unlock")
		unlock := mutex.Unlock(timeout)
		log.Println("unlock结果：", unlock)
	}()

}
