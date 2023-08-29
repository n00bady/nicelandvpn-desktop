package core

import (
	"net"
)

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func BUILD_NAT_MAP(AP *AccessPoint) (err error) {
	AP.NAT_CACHE = make(map[[4]byte][4]byte)
	AP.REVERSE_NAT_CACHE = make(map[[4]byte][4]byte)

	// CreateLog("NAT", "Building map for: ", AP.IP)
	// firstLog := false
	for _, v := range AP.NAT {
		// c := PCIDR(v.LocalNetwork)
		ip2, _, err := net.ParseCIDR(v.Nat)
		if err != nil {
			return err
		}
		ip, ipnet, err := net.ParseCIDR(v.Network)
		if err != nil {
			return err
		}
		firstNetworkByte := ip2.To4()[0]
		// log.Println(ip.To4()[0])
		// log.Println(ipnet)
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {

			AP.NAT_CACHE[[4]byte{firstNetworkByte, ip[1], ip[2], ip[3]}] = [4]byte{ip[0], ip[1], ip[2], ip[3]}

			AP.REVERSE_NAT_CACHE[[4]byte{ip[0], ip[1], ip[2], ip[3]}] = [4]byte{firstNetworkByte, ip[1], ip[2], ip[3]}

			// if !firstLog {
			// log.Println("NAT", "E: >>", [4]byte{firstNetworkByte, ip[1], ip[2], ip[3]}, ">", ip[0], ip[1], ip[2], ip[3])

			// log.Println("NAT", "I: >>", [4]byte{ip[0], ip[1], ip[2], ip[3]}, ">", firstNetworkByte, ip[1], ip[2], ip[3])
			// }
			// firstLog = true
		}

	}
	return
}
