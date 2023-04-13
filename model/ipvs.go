package model

type Ipvs struct{
	VIP string `json:"vip"`
	Backends []*Backend `json:"backends"`
	Protocol string `json:"protocol"`
	SchedName string `json:"sched_name"`
}

type Backend struct{
	Addr string `json:"addr"`
	Weight int `json:"weight"`
	Status int `json:"status"` // ipvs后端的健康状态，1为不健康，0为健康
}

type IpvsList struct{
	IpvsList []*Ipvs `json:"ipvs_list"`
	Json string `json:"-"`
}

// 请求raft的数据结构
type RaftStatsData struct {
	AppliedIndex             string `json:"applied_index"`
	CommitIndex              string `json:"commit_index"`
	FsmPending               string `json:"fsm_pending"`
	LastContact              string `json:"last_contact"`
	LastLogIndex             string `json:"last_log_index"`
	LastLogTerm              string `json:"last_log_term"`
	LastSnapshotIndex        string `json:"last_snapshot_index"`
	LastSnapshotTerm         string `json:"last_snapshot_term"`
	LatestConfiguration      string `json:"latest_configuration"`
	LatestConfigurationIndex string `json:"latest_configuration_index"`
	NumPeers                 string `json:"num_peers"`
	ProtocolVersion          string `json:"protocol_version"`
	ProtocolVersionMax       string `json:"protocol_version_max"`
	ProtocolVersionMin       string `json:"protocol_version_min"`
	SnapshotVersionMax       string `json:"snapshot_version_max"`
	SnapshotVersionMin       string `json:"snapshot_version_min"`
	State                    string `json:"state"`
	Term                     string `json:"term"`
}
type RaftStatsResp struct {
	Data    RaftStatsData   `json:"data"`
	Message string `json:"message"`
}
