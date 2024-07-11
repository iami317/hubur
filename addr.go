package hubur

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ParseStringToPorts 这个函数不要用在最终产品中，仅方便测试，因为可能撑爆内存
func ParseStringToPorts(ports string) []int {
	lports := []int{}

	if strings.HasPrefix(ports, "-") {
		ports = "1" + ports
	}

	if strings.HasSuffix(ports, "-") {
		ports += "65535"
	}

	for _, raw := range strings.Split(ports, ",") {
		raw = strings.TrimSpace(raw)
		if strings.Contains(raw, "-") {
			var (
				low  int64
				high int64
				err  error
			)
			portRange := strings.Split(raw, "-")

			low, err = strconv.ParseInt(portRange[0], 10, 32)
			if err != nil {
				continue
			}

			if portRange[1] != "" {
				high, err = strconv.ParseInt(portRange[1], 10, 32)
				if err != nil {
					continue
				}
			} else {
				continue
			}

			if low > high {
				continue
			}

			for i := low; i <= high; i++ {
				lports = append(lports, int(i))
			}
		} else {
			port, err := strconv.ParseInt(raw, 10, 32)
			if err != nil {
				continue
			}
			lports = append(lports, int(port))
		}
	}

	sort.Ints(lports)
	return lports
}

// ParseStringToHosts 这个函数不要用在最终产品中，仅方便测试！！
// 因为可能撑爆内存
func ParseStringToHosts(raw string) []string {
	targets := []string{}
	for _, h := range strings.Split(raw, ",") {
		// 解析 IP
		if ret := net.ParseIP(FixForParseIP(h)); ret != nil {
			targets = append(targets, ret.String())
			continue
		}

		// 解析 CIDR 网段
		_ip, netBlock, err := net.ParseCIDR(h)
		if err != nil {
			if strings.Count(h, "-") == 1 {
				// 这里开始解析 1.1.1.1-3 的情况
				rets := strings.Split(h, "-")

				// 检查第一部分是不是 IP 地址
				var startIP net.IP
				if startIP = net.ParseIP(rets[0]); startIP == nil {
					targets = append(targets, h)
					continue
				}

				if strings.Count(rets[0], ".") == 3 {
					ipBlocks := strings.Split(rets[0], ".")
					startInt, err := strconv.ParseInt(ipBlocks[3], 10, 64)
					if err != nil {
						targets = append(targets, h)
						continue
					}

					endInt, err := strconv.ParseInt(rets[1], 10, 64)
					if err != nil {
						targets = append(targets, h)
						continue
					}

					if (endInt > 256) || endInt < startInt {
						targets = append(targets, h)
						continue
					}

					additiveRange := endInt - startInt
					low, err := IPv4ToUint32(startIP.To4())
					if err != nil {
						targets = append(targets, h)
						continue
					}

					for i := 0; i <= int(additiveRange); i++ {
						_ip := Uint32ToIPv4(uint32(i) + low)
						if _ip != nil {
							targets = append(targets, _ip.String())
						}
					}
				} else {
					targets = append(targets, h)
					continue
				}
			} else {
				targets = append(targets, h)
			}
			continue
		}

		// 如果是 IPv6 的网段，暂不处理
		if _ip.To4() == nil {
			targets = append(targets, h)
			continue
		}

		// 把 IPv4 专成 int
		low, err := IPv4ToUint32(netBlock.IP)
		if err != nil {
			targets = append(targets, h)
			continue
		}

		for i := low; true; i++ {
			_ip := Uint32ToIPv4(i)
			if netBlock.Contains(_ip) {
				targets = append(targets, _ip.String())
			} else {
				break
			}
		}
	}

	return targets
}

func ParseIPv6(ipString string) net.IP {
	if ip := net.ParseIP(ipString); ip != nil {
		return ip
	}
	return nil
}

func IPv4ToUint32(ip net.IP) (uint32, error) {
	if len(ip) == 4 {
		return binary.BigEndian.Uint32(ip), nil
	} else {
		return 0, fmt.Errorf("cannot convert for ip is not ipv4 ip byte len: %d", len(ip))
	}
}

func Uint32ToIPv4(ip uint32) net.IP {
	ipAddr := make([]byte, 4)
	binary.BigEndian.PutUint32(ipAddr, ip)
	return ipAddr
}

func ParseHostToAddrString(host string) string {
	ip := net.ParseIP(host)
	if ip == nil {
		return host
	}

	if ret := ip.To4(); ret == nil {
		return fmt.Sprintf("[%v]", ip.String())
	}

	return host
}

// PublicIpInfo public ip info: country, region, isp, city, lat, lon, ip
type PublicIpInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Ip          string  `json:"query"`
}

// GetPublicIpInfo return public ip information
// return the PublicIpInfo struct
func GetPublicIpInfo() (*PublicIpInfo, error) {
	resp, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ip PublicIpInfo
	err = json.Unmarshal(body, &ip)
	if err != nil {
		return nil, err
	}

	return &ip, nil
}

// GetIps return all ipv4 of system
func GetIps() []string {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		ipNet, isValid := addr.(*net.IPNet)
		if isValid && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

// GetMacAddrs get mac address
func GetMacAddrs() []string {
	var macAddrs []string

	nets, err := net.Interfaces()
	if err != nil {
		return macAddrs
	}

	for _, net := range nets {
		macAddr := net.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}
		macAddrs = append(macAddrs, macAddr)
	}

	return macAddrs
}

func HostPort(host string, port interface{}) string {
	return fmt.Sprintf("%v:%v", ParseHostToAddrString(host), port)
}

func FixForParseIP(host string) string {
	// 如果传入了 [::] 给 net.ParseIP 则会失败...
	// 所以这里要特殊处理一下
	if strings.Count(host, ":") >= 2 {
		if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
			return host[1 : len(host)-1]
		}
	}
	return host
}

// IsIP 验证是否是合法的 ip 地址
func IsIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsPort checks if a string represents a valid port
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && !strings.Contains(str, ":")
}

func IsIPv6(raw string) bool {
	if ip := net.ParseIP(raw); ip != nil {
		return ip.To4() == nil
	}
	return false
}

func IsPrivate(str string) bool {
	return net.ParseIP(str).IsPrivate()
}

func isDomain(domain string) bool {
	var rege = regexp.MustCompile(`[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+.?`)
	if net.ParseIP(domain) != nil {
		return false
	}
	if _, err := ipRangeParse(domain); err == nil {
		return false
	}

	return rege.MatchString(domain)
}

func IsDomain(input string) bool {
	u, err := url.ParseRequestURI(fmt.Sprintf("http://%s", input))
	if err != nil {
		return false
	}
	r, _ := regexp.Compile("^[a-zA-Z0-9]+([\\-\\.][a-zA-Z0-9]+)*\\.[a-zA-Z]{2,}$")
	if r.MatchString(u.Host) {
		return true
	} else {
		return false
	}
}

// CheckIfASN checks if the given input is ASN or not,
// its possible to have an org name starting with AS/as prefix.
func CheckIfASN(input string) bool {
	if len(input) == 0 {
		return false
	}
	hasASNPrefix := stringsutil.HasPrefixI(input, "AS")
	if hasASNPrefix {
		input = input[2:]
	}
	return hasASNPrefix && CheckIfASNId(input)
}

func CheckIfASNId(input string) bool {
	if len(input) == 0 {
		return false
	}
	hasNumericId := input != "" && govalidator.IsNumeric(input)
	return hasNumericId
}

// IsPublicIP verify a ip is public or not
func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
