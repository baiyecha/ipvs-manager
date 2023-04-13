package conf


type ConfigRaft struct {
	NodeId     string `mapstructure:"node_id"`
	VolumeDir  string `mapstructure:"volume_dir"`
	ClusterAddress []string `mapstructure:"cluster_address"`
	ClusterAdvertise string `mapstructure:"cluster_advertise"`
}

// configServer configuration for HTTP server
type ConfigServer struct {
	Port int `mapstructure:"port"`
	ClusterAddress []string `mapstructure:"cluster_address"`
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
