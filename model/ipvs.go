package model

type Ipvs struct {
	VIP       string     `json:"vip"`
	Backends  []*Backend `json:"backends"`
	Protocol  string     `json:"protocol"`
	SchedName string     `json:"sched_name"`
}

type Backend struct {
	Addr      string `json:"addr"`
	Weight    int    `json:"weight"`
	Status    int    `json:"status"`      // ipvs后端的健康状态，1为不健康，0为健康
	CheckType int    `json:"check_type"`  // 0 为tcp 1为http
	CheckInfo string `json:"check_info"` // 检查的地址，如果type是tcp，那么使用tcp检查，这里为空的时候用addr进行，如果是http，这这里必须为一个可以get的http的地址
}

type IpvsList struct {
	List []*Ipvs `json:"list"`
	Json     string  `json:"-"`
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
	Data    RaftStatsData `json:"data"`
	Message string        `json:"message"`
}
