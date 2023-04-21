package main

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
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
	grpcAddress      = "GRPC_ADDRESS" // agent对接的grpc地址列表
	clusterAddress   = "CLUSTER"      // 集群所有节点的http地址，用对接raft
	clusterAdvertise = "ADVERTIES"    // 集群raft广播出来的地址，集群之间用这个地址通信
	grpcPort         = "GRPC_PORT"    // grpc的监听地址
	dummyName = "DUMMY_NAME" // ipvs网卡的名字
)

var confKeys = []string{
	serverType,
	serverPort,
	raftNodeId,
	raftVolDir,
	grpcAddress,
	clusterAddress,
	clusterAdvertise,
	dummyName,
}

// main entry point of application start
// run using CONFIG=config.yaml ./program
func main() {
	// v := viper.New()
	// v.AutomaticEnv()
	// if err := v.BindEnv(confKeys...); err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	// 使用命令行解析
	pflag.String("server_type", "", "启动方式，默认是只启动server服务编辑ipvs策略, singleon 为all-in-one模式，agent为部署agent控制ipvs")
	pflag.Int("server_port", 8010, "http的端口服务")
	pflag.String("raft_node_id", "raft", "raft 的节点id,每个节点需要保持唯一")
	pflag.String("raft_vol_dir", "node_1_data", "raft⋅信息和kv数据库的文件目录")
	pflag.String("grpc_address", "127.0.0.1:8210", "agent对接的grpc地址列表")
	pflag.String("cluster", "127.0.0.1:8110", "集群所有节点的http地址，用来对接raft")
	pflag.String("adverties", "127.0.0.1:8110", "集群raft广播出来的地址，集群之间用这个地址通信")
	pflag.Int("grpc_port", 8210, "grpc的监听地址")
	pflag.String("dummy_name", "ipvs-manager", "ipvs dummy网卡的名字")
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()
	cluster := strings.Split(viper.GetString(clusterAddress), ",")
	conf := conf.Config{
		Server: conf.ConfigServer{
			Port:           viper.GetInt(serverPort),
			ClusterAddress: cluster,
		},
		Raft: conf.ConfigRaft{
			NodeId:           viper.GetString(raftNodeId),
			VolumeDir:        viper.GetString(raftVolDir),
			ClusterAddress:   cluster,
			ClusterAdvertise: viper.GetString(clusterAdvertise),
		},
		Agent: conf.AgentConf{
			GrpcAddress: strings.Split(viper.GetString(grpcAddress), ","),
			DummtName: viper.GetString(dummyName),
		},
		Grpc: conf.GrpcConf{
			Port: viper.GetInt(grpcPort),
		},
	}

	log.Printf("%+v\n", conf)
	switch viper.GetString(serverType) {
	case "singleon": // all-in-one

	case "agent": // 单agent
		ipvsAgent.RunAgent(conf.Agent)
	default: // 默认启动server
		server.NewServer(conf.Raft, conf.Server.Port)

	}
}
