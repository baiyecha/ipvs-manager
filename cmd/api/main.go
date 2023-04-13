package main

import (
	"log"
	"strings"

	"github.com/spf13/viper"

	"baiyecha/ipvs-manager/conf"
	"baiyecha/ipvs-manager/ipvsAgent"
	"baiyecha/ipvs-manager/server"
)

// configRaft configuration for raft node

const (
	serverType       = "SERVER_TYPE"  // 服务是何类型根据类型启动不同的功能
	serverPort       = "SERVER_PORT"  // http服务的端口
	raftNodeId       = "RAFT_NODE_ID" // node的id
	raftVolDir       = "RAFT_VOL_DIR" // raft 信息和kv数据库的文件目录
	grpcAddress      = "GRPC_ADDREDD" // agent对接的grpc地址列表
	clusterAddress   = "CLUSTER"      // 集群所有节点的http地址，用对接raft
	clusterAdvertise = "ADVERTIES"    // 集群raft广播出来的地址，集群之间用这个地址通信
)

var confKeys = []string{
	serverType,
	serverPort,
	raftNodeId,
	raftVolDir,
	grpcAddress,
	clusterAddress,
	clusterAdvertise,
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
	cluster := strings.Split(v.GetString(clusterAddress), ",")
	conf := conf.Config{
		Server: conf.ConfigServer{
			Port:           v.GetInt(serverPort),
			ClusterAddress: cluster,
		},
		Raft: conf.ConfigRaft{
			NodeId:           v.GetString(raftNodeId),
			VolumeDir:        v.GetString(raftVolDir),
			ClusterAddress:   cluster,
			ClusterAdvertise: v.GetString(clusterAdvertise),
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
		server.NewServer(conf.Raft, conf.Server.Port)

	}
}
