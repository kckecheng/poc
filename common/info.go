package common

import (
	"net"
	"os"
	"runtime"

	"github.com/google/uuid"
)

// BasicInfo basic OS level information
type BasicInfo struct {
	UUID     string
	Hostname string
	Arch     string
	OS       string
	Address  map[string][]string // interface as key, ip as values
}

// NewBasicInfo init basic info
func NewBasicInfo() BasicInfo {
	return BasicInfo{}.
		withUUID().
		withHostname().
		withArch().
		withOS().
		withAddresses()
}

func (binfo BasicInfo) withUUID() BasicInfo {
	binfo.UUID = uuid.New().String()
	return binfo
}

func (binfo BasicInfo) withHostname() BasicInfo {
	hname, err := os.Hostname()
	if err != nil {
		hname = "unknown"
	}
	binfo.Hostname = hname
	return binfo
}

func (binfo BasicInfo) withArch() BasicInfo {
	binfo.Arch = runtime.GOARCH
	return binfo
}

func (binfo BasicInfo) withOS() BasicInfo {
	binfo.OS = runtime.GOOS
	return binfo
}

func (binfo BasicInfo) withAddresses() BasicInfo {
	addrs := map[string][]string{}

	ifaces, err := net.Interfaces()
	if err != nil {
		binfo.Address = addrs
		return binfo // interfaces cannot be identified, hance just return an empty map
	}

	for _, iface := range ifaces {
		ifname := iface.Name
		addrs[ifname] = []string{}

		ips, err := iface.Addrs()
		if err != nil {
			continue // if ip cannot be idenified, just skip the iface
		}
		for _, ip := range ips {
			addrs[ifname] = append(addrs[ifname], ip.String())
		}
	}
	binfo.Address = addrs
	return binfo
}
