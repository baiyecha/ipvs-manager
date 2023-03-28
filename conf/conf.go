package conf


type ConfigRaft struct {
	NodeId     string `mapstructure:"node_id"`
	Port       int    `mapstructure:"port"`
	VolumeDir  string `mapstructure:"volume_dir"`
	RaftLeader string `mapstructure:"raft_leader"`
}

// configServer configuration for HTTP server
type ConfigServer struct {
	Port int `mapstructure:"port"`
}
type AgentConf struct {
	GrpcAddress []string `mapstructure:"grpc_address"`
}

// config configuration
type Config struct {
	Server ConfigServer `mapstructure:"server"`
	Raft   ConfigRaft   `mapstructure:"raft"`
	Agent  AgentConf    `mapstructure:"agent"`
}
