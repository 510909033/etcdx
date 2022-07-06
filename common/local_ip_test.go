package common

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	ip, err := GetLocalIP()
	assert.Nil(t, err)
	log.Println(ip, err)
	t.Log("GetLocalIP=", ip)
}

func TestGetLocalIPList(t *testing.T) {
	ipList, err := GetLocalIPList()
	assert.Nil(t, err)
	t.Logf("GetLocalIPList返回的IP列表为：%+v\n", ipList)
}
