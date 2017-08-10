package traefikZookeeper

import (
	"errors"
	"net"
)

func DefaultIPAddress() (string, error) {
	addr, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}
	if len(addr) == 0 {
		return "", errors.New("No network device was found.")
	}

	var ip string
	switch v := addr[0].(type) {
	case *net.IPNet:
		ip = v.IP.String()
	case *net.IPAddr:
		ip = v.IP.String()
	}

	return ip, nil
}

func IPAddressFromIface(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)

	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()

	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", errors.New("No address was found.")
	}

	var ip string
	switch v := addrs[0].(type) {
	case *net.IPNet:
		ip = v.IP.String()
	case *net.IPAddr:
		ip = v.IP.String()
	}

	return ip, nil
}
