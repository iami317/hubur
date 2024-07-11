package hubur

import (
	"context"
	"errors"
	"fmt"
	"github.com/kataras/golog"
	"github.com/malfunkt/iprange"
	"github.com/projectdiscovery/cdncheck"
	"net"
	"sync"
	"time"
)

type ParseResult struct {
	net.IP
	DomainName string
}

func (r *ParseResult) String() string {
	if r.DomainName != "" {
		return r.IP.String() + ":" + r.DomainName
	} else {
		return r.IP.String()
	}
}

type ParseTarget struct {
	log             *golog.Logger
	ctx             context.Context
	Target          []string
	DNSServer       []*net.Resolver
	Whitelist       []string
	filterCdnClient *cdncheck.Client
	filterIpv6      bool //是否跳过ipv6
}

// 启动配置参数
type ParseDomainOptions struct {
	Ctx        context.Context
	Log        *golog.Logger
	Target     []string
	DnsServer  []string //dns
	Whitelist  []string //过滤白名单
	FilterCdn  bool     //过滤vpn
	FilterIpv6 bool     //过滤ipv6
}

func NewParseTarget(options ParseDomainOptions) (*ParseTarget, error) {
	var r []*net.Resolver
	var cdnCheckClient *cdncheck.Client
	//var err error
	if options.DnsServer != nil {
		for _, dns := range options.DnsServer {
			ip := net.ParseIP(dns)
			if ip == nil {
				return nil, errors.New("DNS server addr parse error(IP format error)")
			}
			resolve := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: 10 * time.Second,
					}
					return d.DialContext(ctx, "udp", fmt.Sprintf("%v:%v", dns, 53))
				},
			}
			r = append(r, resolve)
		}
	}
	if options.FilterCdn {
		cdnCheckClient = cdncheck.New()
	}
	return &ParseTarget{
		Target:          options.Target,
		DNSServer:       r,
		Whitelist:       options.Whitelist,
		ctx:             options.Ctx,
		log:             options.Log,
		filterCdnClient: cdnCheckClient,
		filterIpv6:      options.FilterIpv6,
	}, nil
}

func (p *ParseTarget) ParseIp() ([]ParseResult, error) {
	targetList, err := p.parse(p.Target)
	if err != nil {
		return nil, err
	}
	whiteList, err := p.parse(p.Whitelist)
	if err != nil {
		return nil, err
	}
	targetList = RemoveWriteList(targetList, whiteList)
	return RemoveRepByMap(targetList), nil
}

func RemoveWriteList(target []ParseResult, writeList []ParseResult) []ParseResult {
	for _, w := range writeList {
		target = DeleteSliceIP(target, w)
	}
	return target
}

func DeleteSliceIP(a []ParseResult, r ParseResult) []ParseResult {
	j := 0
	for _, val := range a {
		if val.IP.String() != r.String() && val.String() != r.String() {
			a[j] = val
			j++
		}
	}
	return a[:j]
}

func (p *ParseTarget) parse(target []string) ([]ParseResult, error) {
	var result []ParseResult
	var lock sync.RWMutex
	var Thread = 50
	var err error
	listChan := make(chan string, 100)
	var wg sync.WaitGroup

	//启动50线程用于解析域名
	for i := 0; i < Thread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for l := range listChan {
				var iplist []string
				if p.DNSServer != nil { //是否指定dns服务器
					for _, dnsServer := range p.DNSServer { //根据dns服务器顺序做解析
						iplist, err = dnsServer.LookupHost(p.ctx, l)
						if err != nil {
							p.log.Info(err) //解析失败使用下一个
							continue
						} else {
							break //解析成功退出
						}
					}
				} else { //如果没指定默认使用本机的
					iplist, err = net.LookupHost(l)
					if err != nil {
						p.log.Info(err)
					}
				}
				if iplist != nil { //解析成功添加到结果列表中
					for _, ip := range iplist {
						//判断是否是cdn
						netip := net.ParseIP(ip)
						if p.filterCdnClient != nil {
							if found, _, _, err := p.filterCdnClient.Check(netip); found && err == nil {
								p.log.Infof("domain %v is resolve to %v ,It's cnd automatically skip", l, ip)
								continue //检查出来是cdn直接跳过
							}
						}

						if p.filterIpv6 { //判断ipv6 过滤
							if IsIPv6(netip.String()) {
								p.log.Infof("domain %v is ipv6 to %v ,It's skip", l, ip)
								continue
							}
						}
						lock.Lock()
						result = append(result, ParseResult{netip, l})
						lock.Unlock()
					}
				} else {
					p.log.Infof("Domain name resolution failed:%v", l) //解析结果为空，说明所有dns服务器对当前域名解析都失败
				}
				select {
				case <-p.ctx.Done():
					return
				default:
					continue
				}
			}
		}()
	}
	for _, v := range target {
		ip, err := ipRangeParse(v)
		if err != nil {
			if isDomain(v) {
				listChan <- v
			} else {
				return nil, err
			}
		} else {
			result = append(result, ip...)
		}
	}
	close(listChan)
	wg.Wait()
	return result, nil
}

// 解析ip段172.16.95.1/24  172.16.95.1-40 172.16.95-100.1-40
func ipRangeParse(ip string) ([]ParseResult, error) {
	var re []ParseResult
	list, err := iprange.ParseList(ip)
	if err != nil { //解析不了的ip先判断是不是ipv6
		if IsIPv6(ip) {
			return []ParseResult{{ParseIPv6(ip), ""}}, nil
		}
		return nil, fmt.Errorf("IP format error,check the entered IP address:%v", ip)
	}
	iplist := list.Expand()
	for _, i := range iplist {
		re = append(re, ParseResult{i, ""})
	}
	return re, nil
}

// 特殊情况处理，如果存在相同的ip会进行去重
func RemoveRepByMap(slc []ParseResult) []ParseResult {
	result := []ParseResult{}
	tempMap := map[string]string{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		do, _ := tempMap[e.String()]
		tempMap[e.String()] = e.DomainName
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		} else if len(tempMap) == l && e.DomainName != do { //ip虽然相同但是域名不同也保留
			result = append(result, e)
		}
	}
	return result
}
