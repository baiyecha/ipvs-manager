package utils

import (
	"baiyecha/ipvs-manager/model"
	"fmt"

	"github.com/hashicorp/raft"
	"github.com/levigross/grequests"
)

func GetLeader(address []string) string {
	// fmt.Println("address", address)
	leaderAddr := ""
	for _, addr := range address {
		res, err := grequests.Get(fmt.Sprintf("http://%s/raft/stats", addr), &grequests.RequestOptions{})
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp := &model.RaftStatsResp{}
		if err = res.JSON(resp); err != nil {
			fmt.Println(err)
			continue
		}
		// fmt.Printf("GetLeader res: status: %d , data: %+v \n",res.StatusCode, resp)
		if resp.Data.State == raft.Leader.String() {
			// fmt.Println("leader is ", addr)
			leaderAddr = addr
			break
		}
	}
	return leaderAddr
}
