package ipvsAgent

import (
	"fmt"
	"runtime/debug"
	"time"

	"baiyecha/ipvs-manager/conf"
	"baiyecha/ipvs-manager/grpc/client"
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
				ipvsList, err := getIpvs(agentConf.GrpcAddress, agentConf.AgentAdvertise)
				if err != nil {
					fmt.Print("any addr is connection fail")
					time.Sleep(5 * time.Second)
					continue
				}
				if ipvsList == nil {
					time.Sleep(5 * time.Second)
					continue
				}
				HandleIpvs(ipvsList, agentConf.DummtName)
				time.Sleep(5 * time.Second)
			}
		}()
		time.Sleep(5 * time.Second)
	}
}

func getIpvs(address []string, agentAdvertise string) (ipvsList *model.IpvsList, err error) {
	c := client.NewGrpClient(address...)
	ipvsList, err = c.GetIpvsList(agentAdvertise)
	if err != nil {
		return nil, err
	}
	if ipvsList == nil {
		return nil, fmt.Errorf("get ipvslist is nil")
	}
	return ipvsList, err
}
