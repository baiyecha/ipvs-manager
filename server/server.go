package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

	"baiyecha/ipvs-manager/constant"
	pb "baiyecha/ipvs-manager/grpc/proto"
	"baiyecha/ipvs-manager/model"

	"baiyecha/ipvs-manager/conf"
	"baiyecha/ipvs-manager/fsm"
	"baiyecha/ipvs-manager/server/store_handler"

	"baiyecha/ipvs-manager/utils"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)

func NewRaft(conf conf.ConfigRaft, port int, db *badger.DB) (*raft.Raft, error) {
	_, raftBinAddr, err := net.SplitHostPort(conf.ClusterAdvertise)
	if err != nil {
		panic(err)
	}
	raftBinAddr = fmt.Sprintf(":%s", raftBinAddr)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.NodeId)
	raftConf.SnapshotThreshold = 20
	raftConf.SnapshotInterval = 60 * time.Second
	raftConf.LogLevel = "warn"

	fsmStore := fsm.NewBadger(db)

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
	leraderAddr := utils.GetLeader(conf.RaftListenPeer)
	if leraderAddr == "" { // 如果是空，则以leader启动，否则以follower身份加入集群
		raftServer.BootstrapCluster(configuration)
	} else {
		err := joinRaftCluster(conf.NodeId, string(configuration.Servers[0].Address), leraderAddr)
		if err != nil {
			panic(err)
		}
	}
	return raftServer, err
}

// port is http port
func NewServer(conf conf.ConfigRaft, port int, grpcConf conf.GrpcConf, c conf.Config) {
	badgerOpt := badger.DefaultOptions(conf.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()
	raftServer, err := NewRaft(conf, port, badgerDB)
	if err != nil {
		panic(err)
	}
	// 开始心跳检测
	go RunHealthCheck(badgerDB, raftServer)
	go newGrpcServer(grpcConf, badgerDB, conf.RaftListenPeer)
	go heartbeat(badgerDB, c, raftServer)
	// 开启http服务
	NewHttp(fmt.Sprintf(":%d", port), fmt.Sprintf(":%d", conf.RaftHttpPort), badgerDB, raftServer, conf.RaftListenPeer)
}

func heartbeat(db *badger.DB, conf conf.Config, r *raft.Raft) {
	ip, _, _ := net.SplitHostPort(conf.Raft.ClusterAdvertise)
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("heartbeat task error", "err", r, "stack", string(debug.Stack()))
				}
			}()
			for {
				// 定时请求进行心跳检测
				store_handler.Heartbeat(db, &model.NodeInfo{
					IP:       ip,
					RpcPort:  strconv.Itoa(conf.Grpc.Port),
					WebPort:  strconv.Itoa(conf.Server.Port),
					IsLeader: strconv.FormatBool(r.State() == raft.Leader),
				}, constant.ServerRule, conf.Server.RaftListenPeer)
				time.Sleep(30 * time.Second)
			}
		}()
	}
}

func newGrpcServer(conf conf.GrpcConf, db *badger.DB, clusterAddress []string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}
	s := grpc.NewServer()
	pb.RegisterIpvsListServiceServer(s, store_handler.NewGrpcStoreServer(db, clusterAddress))
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return err
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
