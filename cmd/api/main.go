package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/spf13/viper"

	"ysf/raftsample/fsm"
	"ysf/raftsample/server"
)

// configRaft configuration for raft node
type configRaft struct {
	NodeId     string `mapstructure:"node_id"`
	Port       int    `mapstructure:"port"`
	VolumeDir  string `mapstructure:"volume_dir"`
	RaftLeader string `mapstructure:"raft_leader"`
}

// configServer configuration for HTTP server
type configServer struct {
	Port int `mapstructure:"port"`
}

// config configuration
type config struct {
	Server configServer `mapstructure:"server"`
	Raft   configRaft   `mapstructure:"raft"`
}

const (
	serverPort = "SERVER_PORT"  // http服务的端口
	raftNodeId = "RAFT_NODE_ID" // node的id
	raftPort   = "RAFT_PORT"    // raft监听端口
	raftVolDir = "RAFT_VOL_DIR" // raft 信息和kv数据库的文件目录
	raftLeader = "RAFT_LEADER"  // 如果是leader，则为空，如果是follower,则需要填leader的节点信息
)

var confKeys = []string{
	serverPort,
	raftNodeId,
	raftPort,
	raftVolDir,
	raftLeader,
}

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

// main entry point of application start
// run using CONFIG=config.yaml ./program
func main() {
	v := viper.New()
	v.AutomaticEnv()
	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	conf := config{
		Server: configServer{
			Port: v.GetInt(serverPort),
		},
		Raft: configRaft{
			NodeId:     v.GetString(raftNodeId),
			Port:       v.GetInt(raftPort),
			VolumeDir:  v.GetString(raftVolDir),
			RaftLeader: v.GetString(raftLeader),
		},
	}

	log.Printf("%+v\n", conf)

	// Preparing badgerDB
	badgerOpt := badger.DefaultOptions(conf.Raft.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()

	raftBinAddr := fmt.Sprintf(":%d", conf.Raft.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)
	raftConf.SnapshotThreshold = 1024

	fsmStore := fsm.NewBadger(badgerDB)

	store, err := raftboltdb.NewBoltStore(filepath.Join(conf.Raft.VolumeDir, "raft.dataRepo"))
	if err != nil {
		panic(err)
	}

	// Wrap the store in a LogCache to improve performance.
 cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		panic(err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, 1, os.Stdout)
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
				ID:      raft.ServerID(conf.Raft.NodeId),
				Address: transport.LocalAddr(),
			},
		},
	}
	if conf.Raft.RaftLeader == "" { // 如果是空，则以leader启动，否则以follower身份加入集群
		raftServer.BootstrapCluster(configuration)
	} else {
		err := joinRaftCluster(conf.Raft.NodeId, string(configuration.Servers[0].Address), conf.Raft.RaftLeader)
		if err != nil{
			panic(err)
		}
	}

	srv := server.New(fmt.Sprintf(":%d", conf.Server.Port), badgerDB, raftServer)
	if err := srv.Start(); err != nil {
		panic(err)
	}

	return
}

func joinRaftCluster(node_id, raft_address, raft_leader string) error{
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

	req, err := http.NewRequest("POST",fmt.Sprintf("http://%s/raft/join", raft_leader) ,body)
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
