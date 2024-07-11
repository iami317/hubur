package hubur

import (
	"context"
	"github.com/kataras/golog"
	"testing"
)

func TestParseTarget_ParseIp(t *testing.T) {
	targetList := []string{
		//"172.16.95.1/28",
		//"172.16.95-100.1-10",
		//"fe80::c7b:402b:959f:7ba1",
		//"fe80::147d:daff:fee3:1f64",
		//"zyylhn.cn",
		"aahzdsysya.co",
		"huaun.com",
		"widget-v4.tidiochat.com",
	}
	writeList := []string{"172.16.95.14", "172.16.97.1-6"}
	dnsServer := []string{
		"192.168.10.255",
		//"114.114.114.114",
		"8.8.8.8",
	}
	options := ParseDomainOptions{
		Ctx:        context.Background(),
		Log:        golog.New(),
		Target:     targetList,
		DnsServer:  dnsServer,
		Whitelist:  writeList,
		FilterCdn:  false,
		FilterIpv6: false,
	}
	target, err := NewParseTarget(options)
	if err != nil {
		t.Error(err)
	}
	re, err := target.ParseIp()
	if err != nil {
		t.Error(err)
	}
	t.Log(re)
}
