package main

import (
	"log"
	"strings"

	"github.com/spf13/viper"

	"ysf/raftsample/server"
	"ysf/raftsample/conf"
	"ysf/raftsample/ipvsAgent"
)

// configRaft configuration for raft node

const (
	serverType  = "SERVER_TYPE"  // 服务是何类型根据类型启动不同的功能
	serverPort  = "SERVER_PORT"  // http服务的端口
	raftNodeId  = "RAFT_NODE_ID" // node的id
	raftPort    = "RAFT_PORT"    // raft监听端口
	raftVolDir  = "RAFT_VOL_DIR" // raft 信息和kv数据库的文件目录
	raftLeader  = "RAFT_LEADER"  // 如果是leader，则为空，如果是follower,则需要填leader的节点信息
	grpcAddress = "GRPC_ADDREDD" // agent对接的grpc地址列表
)

var confKeys = []string{
	serverType,
	serverPort,
	raftNodeId,
	raftPort,
	raftVolDir,
	raftLeader,
	grpcAddress,
}


// main entry point of application start
// run using CONFIG=config.yaml ./program
func main() {
	v := viper.New()
	v.AutomaticEnv()
	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}
	conf := conf.Config{
		Server: conf.ConfigServer{
			Port: v.GetInt(serverPort),
		},
		Raft: conf.ConfigRaft{
			NodeId:     v.GetString(raftNodeId),
			Port:       v.GetInt(raftPort),
			VolumeDir:  v.GetString(raftVolDir),
			RaftLeader: v.GetString(raftLeader),
		},
		Agent: conf.AgentConf{
			GrpcAddress: strings.Split(v.GetString(grpcAddress), ","),
		},
	}

	log.Printf("%+v\n", conf)
	switch v.GetString(serverType) {
	case "singleon": // all-in-one

	case "agent": // 单agent
		ipvsAgent.RunAgent(conf.Agent)
	default: // 默认启动server
	server.NewRaftServer(conf.Raft, conf.Server.Port)

	}
}
