package discovery

import (
	"github.com/510909033/etcdx/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDiscoveryService_host(t *testing.T) {

	name, err := os.Hostname()
	assert.Nil(t, err)
	t.Logf("hostname=%s", name)
}

func TestDiscoveryService_DemoMultiHttpAndRegister(t *testing.T) {

	service := discoveryService{}
	client := common.GetCliV3()
	prefixKey := "/bbt/"
	service.DemoMultiHttpAndRegister(client, prefixKey)
}
