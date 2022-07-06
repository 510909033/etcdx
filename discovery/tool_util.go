package discovery

import (
	"context"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

//常用工具
type ToolsUtil struct {
}

func (tool *ToolsUtil) GetRegisterCount(ctx context.Context, client *clientv3.Client, prefixKey string) int64 {

	response, err := client.Get(ctx, prefixKey, []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithCountOnly(),
	}...)
	if err != nil {
		logrus.Errorf("Get err=%+v", err)
		return 0
	}
	_ = response
	return response.Count

}
