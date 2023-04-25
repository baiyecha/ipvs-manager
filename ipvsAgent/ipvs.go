package ipvsAgent

import (
	"fmt"
	"net"
	"sort"
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
	services, _ := handle.GetServices()
	// 本机的ipvs map
	localIpvsMap := make(map[string]*ipvs.Service)
	for _, s := range services {
		localIpvsMap[s.Address.String()+":"+strconv.Itoa(int(s.Port))] = s
	}
	// 用户配置的ipvs map
	userIpvsMap := make(map[string]*model.Ipvs)
	for _, service := range ipvsList.List {
		userIpvsMap[service.VIP] = service
	}

	// 根据当前的ipvs后端状态，调整权重，如果健康检查失败，那么后端的权重直接为0
	for _, userIpvs := range ipvsList.List {
		for _, backend := range userIpvs.Backends {
			if backend.Status == 1 {
				backend.Weight = 0
			}
		}
	}

	// 对比两个数据，保持一致
	// 先对比需要增加的
	for k, v := range userIpvsMap {
		if _, ok := localIpvsMap[k]; ok {
			// 如果本地存在则跳过
			continue
		}
		// 增加ipvs规则
		createIpvs(v, dummyName)
	}
	// 再对比需要删除的
	for k, v := range localIpvsMap {
		if _, ok := userIpvsMap[k]; ok {
			continue
		}
		// 删除ipvs
		deleteIpvs(v, dummyName)
	}
	// 最后确认后端规则是否有变动，如果有变动，则删除整个ipvs规则再重新生成，保持幂等

	for k, v := range localIpvsMap {
		userIpvs, ok := userIpvsMap[k]
		if !ok {
			continue
		}
		// 遍历双方的backend和转发规则
		// 确认转发规则
		if userIpvs.SchedName != v.SchedName {
			// 更新ipvs调度规则
			updateIpvs(v, userIpvs, dummyName)
		}
		localBackend, _ := handle.GetDestinations(v)
		// 确认backend
		if len(userIpvs.Backends) != len(localBackend) {
			// 更新整个ipvs
			updateIpvs(v, userIpvs, dummyName)
		}
		// 对比backend,
		// 先用ip排个序
		sort.Slice(localBackend, func(i, j int) bool {
			if localBackend[i].Address.String() > localBackend[j].Address.String() {
				return true
			}
			if localBackend[i].Address.String() == localBackend[j].Address.String() {
				return localBackend[i].Port > localBackend[j].Port
			}
			return false
		})
		sort.Slice(userIpvs.Backends, func(i, j int) bool {
			return userIpvs.Backends[i].Addr > localBackend[j].Address.String()
		})
		// 排序完成后，开始对比backend，只要有不同，就直接走更新整个ipvs逻辑
		for i, userBackend := range userIpvs.Backends {
			if userBackend.Weight != localBackend[i].Weight {
				// 更新
				updateIpvs(v, userIpvs, dummyName)
				continue
			}
			host, port, _ := net.SplitHostPort(userBackend.Addr)
			if host != localBackend[i].Address.String() {
				// 更新
				updateIpvs(v, userIpvs, dummyName)
				continue
			}
			if port != strconv.Itoa(int(localBackend[i].Port)) {
				// 更新
				updateIpvs(v, userIpvs, dummyName)
				continue
			}
		}
	}
}

func updateIpvs(Oldservice *ipvs.Service, service *model.Ipvs, dummyName string) error {
	fmt.Println("更新整个ipvs...")
	deleteIpvs(Oldservice, dummyName)
	return createIpvs(service, dummyName)
}

func deleteIpvs(service *ipvs.Service, dummyName string) error {
	handle, err := ipvs.New("")
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer handle.Close()
	checkDummyIface(dummyName)
	err = handle.DelService(service)
	if err != nil {
		fmt.Println("delete ipvs service error", err)
	}
	// 删除ip和iptables
	return delDummyIfaceAddrs(dummyName, []string{service.Address.String()})
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
		SetupIPTables(addr + "/32")
	}
	return nil
}

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
	ipt.Delete("nat", "POSTROUTING", rule...)
	ipt.Append("nat", "POSTROUTING", rule...)
	fmt.Println("Setup IPTables done")
}

func delDummyIfaceAddrs(name string, addrs []string) error {
	// 查看虚拟网卡的信息
	fmt.Sprintln("删除网卡ip", name, addrs)
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
		nladdr := &netlink.Addr{
			IPNet: ipaddr,
			Label: "",
		}
		if err = netlink.AddrDel(link, nladdr); err != nil {
			fmt.Printf("Failed to del IP address: %v", err)
		}
		DelIPTables(addr + "/32")
	}
	return nil
}

func DelIPTables(addr string) {
	// 读取本地iptables
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("iptables new error", err)
		return
	}
	//  iptables -t nat -A POSTROUTING -s 10.0.1.0/24 -j MASQUERADE
	rule := []string{"-s", addr, "-j", "MASQUERADE"}
	// 先进行清除后再添加，保持简易幂等
	ipt.Delete("nat", "POSTROUTING", rule...)
	fmt.Println("del IPTables done")
}
