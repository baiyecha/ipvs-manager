package ipvsAgent

import (
	"fmt"
	"net"
	"strconv"
	"syscall"

	"baiyecha/ipvs-manager/model"

	"github.com/moby/ipvs"
	"github.com/vishvananda/netlink/nl"
)

func HandleIpvs(ipvsList *model.IpvsList) {
	handle, err := ipvs.New("")
	if err != nil {
		fmt.Errorf(err.Error())
	}
	defer handle.Close()
	for _, service := range ipvsList.IpvsList {
		createIpvs(service)
	}
}

func createIpvs(service *model.Ipvs) error {
	handle, err := ipvs.New("")
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer handle.Close()

	svcIp, svcPortStr, err := net.SplitHostPort(service.VIP)
	if err != nil {
		fmt.Println(err)
		return err
	}
	svcPort, _ := strconv.ParseUint(svcPortStr, 10, 16)
	svc := &ipvs.Service{
		Address:       net.ParseIP(svcIp),
		Port:          uint16(svcPort),
		Protocol:      syscall.IPPROTO_TCP,
		AddressFamily: nl.FAMILY_V4,
		Netmask:       0xFFFFFFFF,
		SchedName:     ipvs.RoundRobin,
	}
	fmt.Printf("%+v\n", svc)
	err = handle.NewService(svc)
	if err != nil {
		fmt.Println("44", err)
		return err
	}
	for _, backend := range service.Backends {
		ip, portStr, err := net.SplitHostPort(backend.Addr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		port, _ := strconv.ParseUint(portStr, 10, 16)
		dest := &ipvs.Destination{
			Address: net.ParseIP(ip),
			Port:    uint16(port),
			Weight:  backend.Weight,
		}
		fmt.Printf("%+v\n", dest)
		err = handle.NewDestination(svc, dest)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return err
}
