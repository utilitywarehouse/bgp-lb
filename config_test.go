package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFmt(t *testing.T) {
	c := []byte(`
{
  "bgp": {
    "peers": [
      {
        "address": "10.88.0.253",
	"as": 65512
      },
      {
        "address": "10.88.0.254",
	"as": 65512
      }
    ],
    "local": {
      "routerID": "10.88.0.200",
      "as": 65512,
      "listenPort": -1
    }
  },
  "service": {
    "ip": "10.88.2.1",
    "httphealthcheck": {
      "name": "matchbox",
      "url": "http://localhost:80"
    }

  }
}
`)
	conf := &config{}
	err := json.Unmarshal(c, conf)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(conf.Bgp.Peers))
	assert.Equal(t, "10.88.0.253", conf.Bgp.Peers[0].Address)
	assert.Equal(t, uint32(65512), conf.Bgp.Peers[0].AS)
	assert.Equal(t, "10.88.0.254", conf.Bgp.Peers[1].Address)
	assert.Equal(t, uint32(65512), conf.Bgp.Peers[1].AS)
	assert.Equal(t, "10.88.0.200", conf.Bgp.Local.RouterId)
	assert.Equal(t, uint32(65512), conf.Bgp.Local.AS)
	assert.Equal(t, int32(-1), conf.Bgp.Local.ListenPort)
	assert.Equal(t, "10.88.2.1", conf.Service.IP)
	assert.Equal(t, "matchbox", conf.Service.HttpHealthCheck.Name)
	assert.Equal(t, "http://localhost:80", conf.Service.HttpHealthCheck.Url)
}
