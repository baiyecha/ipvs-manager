package conf

type ConfigRaft struct {
	NodeId           string   `mapstructure:"node_id"`
	VolumeDir        string   `mapstructure:"volume_dir"`
	ClusterAdvertise string   `mapstructure:"cluster_advertise"`
	RaftHttpPort     int      `mapstructure:"raft_http_port"`
	RaftListenPeer   []string `mapstructure:"raft_listen_peer"`
}

// configServer configuration for HTTP server
type ConfigServer struct {
	Port           int      `mapstructure:"port"`
	RaftListenPeer []string `mapstructure:"raft_listen_peer"`
}
type AgentConf struct {
	GrpcAddress []string `mapstructure:"grpc_address"`
	DummtName   string   `mapstructure:"dummy_name"`
}

type GrpcConf struct {
	Port int `mapstructure:"port"`
}

// config configuration
type Config struct {
	Server ConfigServer `mapstructure:"server"`
	Raft   ConfigRaft   `mapstructure:"raft"`
	Agent  AgentConf    `mapstructure:"agent"`
	Grpc   GrpcConf     `mapstructure:"grpc"`
}
