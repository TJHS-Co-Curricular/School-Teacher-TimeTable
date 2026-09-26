// Package netutil 放网络相关的小工具。
package netutil

import "net"

// Addr 是一个本机 IPv4 地址与它的网卡名称。
type Addr struct {
	IP        string
	Interface string
}

// LANAddrs 回传本机在局域网的 IPv4 地址（给同事用）。
func LANAddrs() []Addr {
	var out []Addr
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipn.IP.To4()
			if ip == nil || ip.IsLinkLocalUnicast() {
				continue
			}
			out = append(out, Addr{ip.String(), ifc.Name})
		}
	}
	return out
}

// ClientIP 从 RemoteAddr（"ip:port"）取出 IP。
func ClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// IsLoopback 判断请求是否来自本机。
func IsLoopback(remoteAddr string) bool {
	ip := net.ParseIP(ClientIP(remoteAddr))
	return ip != nil && ip.IsLoopback()
}
