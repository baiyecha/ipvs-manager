package client

import (
	pb "baiyecha/ipvs-manager/grpc/proto"
	"context"

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

func (client *IpvsClient) GetIpvsList() {
}

func doGetIpvsList(address string) (*pb.IpvsListResponse, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure)
	if err != nil {
		return err, nil
	}
	defer conn.Close()
	c := pb.NewIpvsListServiceClient(conn)
	reqBody := &pb.IpvslistRequests{}
	res, err := c.IpvsList(context.Background(), reqBody)
	if err != nil {
		return err, nil
	}
	return nil, res
}
