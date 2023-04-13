package ipvsAgent

import (
	"fmt"
	"runtime/debug"
	"time"

	"baiyecha/ipvs-manager/conf"
	"baiyecha/ipvs-manager/model"
)

func RunAgent(agentConf conf.AgentConf) error {
	fmt.Print("run agent...")
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("consumer task error", "err", r, "stack", string(debug.Stack()))
				}
			}()
			for {
				// 定时请求server端，拿到ipvs信息
				ipvsList, err := getIpvs(agentConf.GrpcAddress)
				if err != nil {
					fmt.Print("any addr is connection fail")
				}
				HandleIpvs(ipvsList)
				time.Sleep(5 * time.Second)
			}
		}()
	}
}

func getIpvs(address []string) (ipvsList *model.IpvsList, err error) {
	ipvsList = &model.IpvsList{
		IpvsList: make([]*model.Ipvs, 0),
	}
	for _, addr := range address {
		// 使用grpc进行通信，获取当前ipvs信息列表
		fmt.Println(addr)

	}
	return ipvsList, err
}
