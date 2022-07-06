package some_demo

import (
	"context"
	"github.com/510909033/etcdx/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

//并发 抢占key=lock的锁
func TestMultiLock(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	for i := 0; i < 5; i++ {
		go lock(3)
	}

	time.AfterFunc(time.Second*50, func() {
		wg.Done()
	})

	wg.Wait()
}

/*
获取锁
如果获取失败， sleep一秒，继续获取
*/
func lock(leaseTTL int64) {
	var client *clientv3.Client
	client = common.GetCliV3()
	var err error

	//上锁并创建租约
	lease := clientv3.NewLease(client)

	var leaseGrantResp *clientv3.LeaseGrantResponse
	if leaseGrantResp, err = lease.Grant(context.TODO(), leaseTTL); err != nil {
		panic(err)
	}
	leaseId := leaseGrantResp.ID
	log.Printf("leaseId=%d\n", leaseId)

	// 创建一个可取消的租约，主要是为了退出的时候能够释放
	ctx, cancelFunc := context.WithCancel(context.TODO())

	// 释放租约
	defer cancelFunc()
	//Revoke撤销给定的租约。
	defer lease.Revoke(context.TODO(), leaseId)

	var keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
	if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		panic(err)
	}
	// 续约应答
	go func() {
		for {
			select {
			case keepResp, ok := <-keepRespChan:
				if !ok {
					log.Println("keepRespChan 通道已关闭")
					return
				}
				if keepRespChan == nil {
					log.Println("租约已经失效了")
					goto END
				} else { // 每秒会续租一次, 所以就会收到一次应答
					log.Println("收到自动续租应答:", keepResp.ID)
				}
			}
		}
	END:
	}()

	// 在租约时间内去抢锁（etcd 里面的锁就是一个 key）
	kv := clientv3.NewKV(client)

	for {
		// 创建事务
		txn := kv.Txn(context.TODO())

		// If 不存在 key，Then 设置它，Else 抢锁失败
		txn.If(clientv3.Compare(clientv3.CreateRevision("lock"), "=", 0)).
			Then(clientv3.OpPut("lock", "g", clientv3.WithLease(leaseId))).
			Else(clientv3.OpGet("lock"))

		// 提交事务
		var txnResp *clientv3.TxnResponse
		if txnResp, err = txn.Commit(); err != nil {
			panic(err)
		}

		if !txnResp.Succeeded {
			log.Println("锁被占用:", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
			time.Sleep(time.Millisecond * 1000)
			continue
		}

		//抢到锁
		break
	}

	// 抢到锁后执行业务逻辑，没有抢到则退出
	log.Println("处理任务")
	time.Sleep(5 * time.Second)

}
