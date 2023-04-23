package client

import (
	"context"
	"fmt"

	pb "baiyecha/ipvs-manager/grpc/proto"
	"baiyecha/ipvs-manager/model"

	"google.golang.org/grpc"
)

type IpvsClient struct {
	address []string
}

func NewGrpClient(address ...string) *IpvsClient {
	return &IpvsClient{
		address: address,
	}
}

func (client *IpvsClient) GetIpvsList() (*model.IpvsList, error) {
	var err error
	for _, addr := range client.address {
		var ipvsListResponse  *pb.IpvsListResponse
		ipvsListResponse, err = doGetIpvsList(addr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ipvsList := transformIpvsList(ipvsListResponse)
		return ipvsList, err
	}
	fmt.Println("do get ipvs by grpc error", err)
	return nil, err
}

func doGetIpvsList(address string) (*pb.IpvsListResponse, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewIpvsListServiceClient(conn)
	reqBody := &pb.IpvsListRequeste{}
	res, err := c.IpvsList(context.Background(), reqBody)
	if err != nil {
		return nil, err
	}
	return res, err
}

func transformIpvsList(ipvsListResponse *pb.IpvsListResponse) *model.IpvsList {
	ipvsList := &model.IpvsList{}
	ipvsList.List = make([]*model.Ipvs, 0, len(ipvsListResponse.List))
	for _, ipvs := range ipvsListResponse.List {
		backends := make([]*model.Backend, 0)
		for _, backend := range ipvs.Backends {
			backends = append(backends, &model.Backend{
				Addr:      backend.Addr,
				Weight:    int(backend.Weight),
				Status:    int(backend.Status),
				CheckType: int(backend.CheckType),
				CheckInfo: backend.CheckInfo,
			})
		}
		ipvsList.List = append(ipvsList.List, &model.Ipvs{
			Backends: backends,
			VIP: ipvs.Vip,
			Protocol: ipvs.Protocol,
			SchedName: ipvs.SchedName,
		})
	}
	return ipvsList
}
