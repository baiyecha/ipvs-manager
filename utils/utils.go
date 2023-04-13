package utils

import (
	"baiyecha/ipvs-manager/model"

	"github.com/hashicorp/raft"
	"github.com/levigross/grequests"
)

func GetLeader(address []string) string {
	leaderAddr := ""
	for _, addr := range address {
		res, err := grequests.Get(addr, &grequests.RequestOptions{})
		if err != nil {
			continue
		}
		resp := &model.RaftStatsResp{}
		if err = res.JSON(resp); err != nil {
			continue
		}
		if resp.Data.State == raft.Leader.String() {
			leaderAddr = addr
			break
		}
	}
	return leaderAddr
}
