package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"
	"ysf/raftsample/conf"
	"ysf/raftsample/fsm"
	"ysf/raftsample/utils"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)
const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)


// port is http port
func NewServer(conf conf.ConfigRaft, port int){
	badgerOpt := badger.DefaultOptions(conf.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()
	_, raftBinAddr, err  := net.SplitHostPort(conf.ClusterAdvertise)
	if err != nil{
		panic(err)
	}

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.NodeId)
	raftConf.SnapshotThreshold = 1024

	fsmStore := fsm.NewBadger(badgerDB)

	store, err := raftboltdb.NewBoltStore(filepath.Join(conf.VolumeDir, "raft.dataRepo"))
	if err != nil {
		panic(err)
	}

	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		panic(err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(conf.VolumeDir, 1, os.Stdout)
	if err != nil {
		panic(err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBinAddr)
	if err != nil {
		panic(err)
	}

	transport, err := raft.NewTCPTransport(raftBinAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		panic(err)
	}

	raftServer, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		panic(err)
	}

	// always start single server as a leader
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(conf.NodeId),
				Address: transport.DecodePeer([]byte(conf.ClusterAdvertise)),
			},
		},
	}
	leraderAddr := utils.GetLeader(conf.ClusterAddress)
	if leraderAddr == "" { // 如果是空，则以leader启动，否则以follower身份加入集群
		raftServer.BootstrapCluster(configuration)
	} else {
		err := joinRaftCluster(conf.NodeId, string(configuration.Servers[0].Address), leraderAddr)
		if err != nil {
			panic(err)
		}
	}

	srv := NewHttp(fmt.Sprintf(":%d", port), badgerDB, raftServer, conf.ClusterAddress)
	if err := srv.Start(); err != nil {
		panic(err)
	}

	return
}

func joinRaftCluster(node_id, raft_address, raft_leader string) error {
	type Payload struct {
		NodeID      string `json:"node_id"`
		RaftAddress string `json:"raft_address"`
	}

	data := Payload{
		NodeID:      node_id,
		RaftAddress: raft_address,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/raft/join", raft_leader), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}
