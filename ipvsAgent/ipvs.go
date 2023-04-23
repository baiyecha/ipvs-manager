package ipvsAgent

import (
	"fmt"
	"net"
	"strconv"
	"syscall"

	"baiyecha/ipvs-manager/model"

	"github.com/coreos/go-iptables/iptables"
	"github.com/moby/ipvs"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

func HandleIpvs(ipvsList *model.IpvsList, dummyName string) {
	handle, err := ipvs.New("")
	if err != nil {
		fmt.Println(err)
	}
	defer handle.Close()
	for _, service := range ipvsList.List {
		createIpvs(service, dummyName)
	}
}

func createIpvs(service *model.Ipvs, dummyName string) error {
	handle, err := ipvs.New("")
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer handle.Close()
	checkDummyIface(dummyName)
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
	vip, _, _ := net.SplitHostPort(service.VIP)
	addDummyIfaceAddrs(dummyName, []string{vip})
	err = handle.NewService(svc)
	if err != nil {
		fmt.Println(err)
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
	// err = addDummyIfaceAddrs(dummyName, []string{service.VIP})
	return err
}

// 检查 dummy 网卡是否存在，不存在就要先创建
func checkDummyIface(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil || link == nil {
		// 网卡不存在，创建网卡
		link := &netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{
				Name: name,
				MTU:  1500,
			},
		}
		if err = netlink.LinkAdd(link); err != nil {
			panic(fmt.Sprintf("Failed to add dummy link: %v", err))
		}
		if err = netlink.LinkSetUp(link); err != nil {
			panic(fmt.Sprintf("Failed to up dummy link: %v", err))
		}
		return err
	}
	netlink.LinkSetUp(link)
	return err
}

func addDummyIfaceAddrs(name string, addrs []string) error {
	// 查看虚拟网卡的信息
	fmt.Sprintln("增加网卡ip", name, addrs)
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get link: %v", err)
	}
	fmt.Printf("Interface name: %s\n", link.Attrs().Name)
	fmt.Printf("Interface hardware address: %s\n", link.Attrs().HardwareAddr.String())
	nladdrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to get address: %v", err)
	}
	nladdrMap := make(map[string]struct{})
	for _, nladdr := range nladdrs {
		fmt.Printf("Interface IP address: %s\n", nladdr.IP.String())
		nladdrMap[nladdr.IP.String()] = struct{}{}
	}
	// 增加ip
	for _, addr := range addrs {
		_, ipaddr, _ := net.ParseCIDR(addr + "/32")
		if err != nil {
			fmt.Println("parse addr error", err, addr)
			continue
		}
		// if _, ok := nladdrMap[addr]; ok {
		// 	continue
		// }
		nladdr := &netlink.Addr{
			IPNet: ipaddr,
			Label: "",
		}
		if err = netlink.AddrAdd(link, nladdr); err != nil {
			if err.Error() != "file exists" {
				fmt.Printf("Failed to add IP address: %v", err)
				continue
			}
		}
		// err = iptables("-t", "nat", "-A", "POSTROUTING", "-s", addr+"/32", "-j", "MASQUERADE")
		//
		//	if err != nil {
		//		fmt.Println("add iptables error ", err, addr)
		//		continue
		//	}
		SetupIPTables(addr + "/32")
	}
	return nil
}

// // # iptables 封装iptables命令
// func iptables(args ...string) error {
// 	fmt.Println("cmd is ", "/sbin/iptables", strings.Join(args, " "))
// 	if err := exec.Command("/sbin/iptables", args...).Run(); err != nil {
// 		return fmt.Errorf("iptables failed: iptables %v", strings.Join(args, " "))
// 	}
// 	return nil
// }

func SetupIPTables(addr string) {
	// 读取本地iptables
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("iptables new error", err)
		return
	}
	//  iptables -t nat -A POSTROUTING -s 10.0.1.0/24 -j MASQUERADE
	rule := []string{"-s", addr, "-j", "MASQUERADE"}
	// 先进行清除后再添加，保持简易幂等
	_ = ipt.Delete("nat", "POSTROUTING", rule...)
	ipt.Append("nat", "POSTROUTING", rule...)
	fmt.Println("Setup IPTables done")
}
